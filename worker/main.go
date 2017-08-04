package worker

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	defer session.Close()
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
	// img := "otiai10/daap-test"
	img := "otiai10/basic-wf"

	env := []string{
		fmt.Sprintf("REFERENCE=%s", "GRCh37.fa"),
	}
	for i, read := range job.Resource.Reads {
		env = append(env, fmt.Sprintf("INPUT%02d=%s", i, read))
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

	err = models.Jobs(session).UpdateId(job.ID, bson.M{
		"$set": bson.M{"stdout": string(out)},
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

	host, err := os.Open(filepath.Join(p, "host.txt"))
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(host)
	if err != nil {
		return nil, err
	}
	return &daap.MachineConfig{
		Host:     strings.Trim(string(buf), " \n"),
		CertPath: filepath.Join(p, "certs"),
	}, nil
}
