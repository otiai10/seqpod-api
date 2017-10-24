package v0

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/otiai10/marmoset"
	"github.com/seqpod/seqpod-api/filters"
	"github.com/seqpod/seqpod-api/models"
	"github.com/seqpod/seqpod-api/worker"
)

// JobWorkspace create job workspace
func JobWorkspace(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	job := models.NewJob()

	// job.Resource.Reference = "GRCh37"
	workflow := r.FormValue("workflow")
	if workflow == "" {
		render.JSON(http.StatusBadRequest, marmoset.P{
			"message": fmt.Errorf("missing required value: `workflow`"),
		})
		return
	}
	job.Workflow = []string{workflow}

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

// JobInputUpload ...
func JobInputUpload(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w)
	sess := filters.MongoSession(r)

	f, h, err := r.FormFile("file")
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

	// Make sure inputs directory exists.
	if err = os.MkdirAll(filepath.Join(job.Resource.URL, "in"), os.ModePerm); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}
	destpath := filepath.Join(job.Resource.URL, "in", h.Filename)

	// Create input physical file.
	destfile, err := os.Create(destpath)
	if err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	// Copy uploaded buffer to physical file.
	if _, err = io.Copy(destfile, f); err != nil {
		render.JSON(http.StatusInternalServerError, marmoset.P{
			"message": err.Error(),
		})
		return
	}

	// }}}

	name := r.FormValue("name")
	// Save uploaded files
	change := bson.M{
		"$set": bson.M{
			fmt.Sprintf("resource.inputs.%s", name): h.Filename,
			"resource.url":                          job.Resource.URL,
		},
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

// Download ...
func Download(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	fname := r.FormValue("result")

	fpath := filepath.Join("/var/app/works", id, fname)

	_, err := os.Stat(fpath)
	if err != nil {
		render := marmoset.Render(w, true)
		render.JSON(http.StatusOK, marmoset.P{
			"id":     r.FormValue("id"),
			"result": r.FormValue("result"),
		})
		return
	}

	f, err := os.Open(fpath)
	if err != nil {
		render := marmoset.Render(w, true)
		render.JSON(http.StatusOK, marmoset.P{
			"id":     r.FormValue("id"),
			"result": r.FormValue("result"),
		})
		return
	}

	io.Copy(w, f)
}
