package worker

import (
	"log"
	"os"
	"time"

	"github.com/otiai10/seqpod-api/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Enqueue enqueues a job to worker queue
// TODO:
//  For now, worker runtime is spawned on a bit worker instance machine, called Elephant.
//  It should be spawned on an instance automatically generated/being terminated by Algnome.
// TODO:
//  For now, worker stdout/stderr is hijacked by worker process.
//  It should be hijacked by DaaP
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

	if err := c.UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"status":     models.Running,
			"started_at": time.Now(),
		},
	}); err != nil {
		failed(session, job, err)
		return
	}

	if err := c.FindId(job.ID).One(job); err != nil {
		failed(session, job, err)
		return
	}

	// {{{ TODO: Exec and wait for DaaP
	time.Sleep(5 * time.Minute)
	// }}}

	if err := c.UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"status":      models.Completed,
			"results":     []string{"foo.txt", "bar.txt"},
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
