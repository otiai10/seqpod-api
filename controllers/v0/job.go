package v0

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/otiai10/marmoset"
	"github.com/otiai10/seqpod-api/filters"
	"github.com/otiai10/seqpod-api/models"
	"github.com/otiai10/seqpod-api/worker"
)

// JobWorkspace create job workspace
func JobWorkspace(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	job := models.NewJob()

	job.Resource.Reference = "GRCh37"
	job.Workflow = []string{"otiai10/basic-wf"}

	if err := models.Jobs(sess).Insert(job); err != nil {
		render.JSON(http.StatusBadRequest, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	render.JSON(http.StatusOK, marmoset.P{
		"job": job,
	})
}

// JobFastqUpload ...
func JobFastqUpload(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	f, h, err := r.FormFile("fastq")
	if err != nil {
		render.JSON(http.StatusBadRequest, marmoset.P{
			"message": err.Error(),
		})
		return
	}
	defer f.Close()

	id := r.FormValue("id")
	job := new(models.Job)
	if err = models.Jobs(sess).FindId(bson.ObjectIdHex(id)).One(job); err != nil {
		if err == mgo.ErrNotFound {
			render.JSON(http.StatusNotFound, marmoset.P{
				"message": err.Error(),
			})
			return
		}
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	// {{{ TODO: Save uploaded files to "/tmp", FOR NOW.
	// it should be something like S3 or any other storage services
	// to make it persistent
	if job.Resource.URL == "" {
		job.Resource.URL = filepath.Join("/var/app/works", id)
	}
	if err = os.MkdirAll(job.Resource.URL, os.ModePerm); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}
	destpath := filepath.Join(job.Resource.URL, h.Filename)
	destfile, err := os.Create(destpath)
	if err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}
	if _, err = io.Copy(destfile, f); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}
	// }}}

	// Save uploaded files
	change := bson.M{
		"$push": bson.M{"resource.reads": h.Filename},
		"$set":  bson.M{"resource.url": job.Resource.URL},
	}
	if err := models.Jobs(sess).UpdateId(bson.ObjectIdHex(id), change); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	render.JSON(http.StatusOK, marmoset.P{
		"job": job,
	})
}

// JobGet ...
func JobGet(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	id := r.FormValue("id")
	job := new(models.Job)

	if err := models.Jobs(sess).FindId(bson.ObjectIdHex(id)).One(job); err != nil {
		render.JSON(http.StatusBadRequest, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	render.JSON(http.StatusOK, marmoset.P{
		"job": job,
	})
}

// JobMarkReady ...
func JobMarkReady(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	id := r.FormValue("id")
	job := new(models.Job)
	if err := models.Jobs(sess).FindId(bson.ObjectIdHex(id)).One(job); err != nil {
		if err == mgo.ErrNotFound {
			render.JSON(http.StatusNotFound, marmoset.P{
				"message": err.Error(),
			})
			return
		}
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	if err := models.Jobs(sess).UpdateId(job.ID, bson.M{
		"$set": bson.M{"status": models.Ready},
	}); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	// ReFetch
	models.Jobs(sess).FindId(job.ID).One(job)
	render.JSON(http.StatusOK, marmoset.P{
		"job": job,
	})

	go worker.Enqueue(job)
}
