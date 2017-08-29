package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/otiai10/daap"
	"github.com/otiai10/seqpod-api/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Enqueue enqueues a job to worker queue
// TODO:
//  For now, worker runtime is spawned on a bit worker instance machine, called "Elephant".
//  It should be spawned on an instance automatically generated/being terminated by "Algnome".
// TODO:
//  For now, result files are placed on API server /tmp directory.
//  It should be placed on S3 Bucket
func Enqueue(job *models.Job) {

	session, err := mgo.Dial(os.Getenv("MONGODB_URI"))
	if err != nil {
		logInternalError("DB SESSION", err)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			failed(session, job, fmt.Errorf("%v", err))
		}
		session.Close()
	}()

	c := models.Jobs(session)

	if err = c.UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"status":     models.Running,
			"started_at": time.Now(),
		},
	}); err != nil {
		failed(session, job, err)
		return
	}

	if err = c.FindId(job.ID).One(job); err != nil {
		failed(session, job, err)
		return
	}

	machine, err := fetchMachineConfig()
	if err != nil {
		failed(session, job, err)
		return
	}
	if len(job.Workflow) == 0 {
		failed(session, job, fmt.Errorf("No any workflow specified"))
		return
	}
	img := job.Workflow[0]

	env := []string{
		fmt.Sprintf("REFERENCE=%s", "GRCh37.fa"),
	}
	for i, read := range job.Resource.Reads {
		env = append(env, fmt.Sprintf("INPUT%02d=%s", i+1, read))
	}

	arg := daap.Args{
		Machine: machine,
		Mounts: []daap.Mount{
			daap.Volume(job.Resource.URL, "/var/data"),
			daap.Volume(job.ReferenceDir(), "/var/refs"),
		},
		Env: env,
	}

	process := daap.NewProcess(img, arg)

	ctx := context.Background()
	if err = process.Run(ctx); err != nil {
		failed(session, job, err)
		return
	}

	out, err := ioutil.ReadAll(process.Stdout)
	if err != nil {
		failed(session, job, err)
		return
	}
	serr, err := ioutil.ReadAll(process.Stderr)
	if err != nil {
		failed(session, job, err)
		return
	}
	applog, err := ioutil.ReadAll(process.Log)
	if err != nil {
		failed(session, job, err)
		return
	}

	err = models.Jobs(session).UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"stdout": string(out),
			"stderr": string(serr),
			"applog": string(applog),
		},
	})
	if err != nil {
		failed(session, job, err)
		return
	}

	// TODO: Use "Salamander"
	results, err := detectResultFiles(job)
	if err != nil {
		failed(session, job, err)
		return
	}

	if err := c.UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"status":      models.Completed,
			"results":     results,
			"finished_at": time.Now(),
		},
	}); err != nil {
		failed(session, job, err)
	}

}

func failed(session *mgo.Session, job *models.Job, err error) {
	if e := models.Jobs(session).UpdateId(job.ID, bson.M{
		"$push": bson.M{
			"errors": err.Error(),
		},
		"$set": bson.M{
			"status":      models.Errored,
			"finished_at": time.Now(),
		},
	}); e != nil {
		logInternalError("Update", e)
	}
}

func logInternalError(prefix string, err error) {
	log.Printf("[WORKER][%s] %v\n", prefix, err.Error())
}

// TODO: Result files are NOT ALWAYS on the root level of directory
//       In future, it should be managed by "Salamander"
//       to place evetything on S3 buckets.
func detectResultFiles(job *models.Job) ([]string, error) {

	inArrayString := func(target string, list []string) bool {
		for _, e := range list {
			if target == e {
				return true
			}
		}
		return false
	}

	files, err := ioutil.ReadDir(job.Resource.URL)
	if err != nil {
		return nil, err
	}
	results := []string{}
	for _, f := range files {
		if inArrayString(f.Name(), job.Resource.Reads) {
			continue
		}
		results = append(results, f.Name())
	}

	return results, nil
}

// fetchMachineConfig fetches machine configs
// from mounted "/var/machine" directory
// so that user can specify machine on docker-compose CLI layer.
// TODO: Machine should be provided by "Algnome",
//       for now, it provides "Elephant"
func fetchMachineConfig() (*daap.MachineConfig, error) {

	// This directory is binded by docker-copose, check docker-copose.yaml.
	p := "/var/machine"

	f, err := os.Open(filepath.Join(p, "config.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	config := new(models.MachineConfig)
	if err := json.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return &daap.MachineConfig{
		Host:     fmt.Sprintf("tcp://%s:2376", config.Driver.IPAddress),
		CertPath: p,
	}, nil
}
