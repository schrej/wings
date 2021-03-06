package router

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pterodactyl/wings/server"
	"github.com/pterodactyl/wings/server/backup"
	"net/http"
)

// Backs up a server.
func postServerBackup(c *gin.Context) {
	s := GetServer(c.Param("server"))

	data := &backup.Request{}
	// BindJSON sends 400 if the request fails, all we need to do is return
	if err := c.BindJSON(&data); err != nil {
		return
	}

	var adapter backup.BackupInterface
	var err error

	switch data.Adapter {
	case backup.LocalBackupAdapter:
		adapter, err = data.NewLocalBackup()
	case backup.S3BackupAdapter:
		adapter, err = data.NewS3Backup()
	default:
		err = errors.New(fmt.Sprintf("unknown backup adapter [%s] provided", data.Adapter))
		return
	}

	if err != nil {
		TrackedServerError(err, s).AbortWithServerError(c)
		return
	}

	go func(b backup.BackupInterface, serv *server.Server) {
		if err := serv.Backup(b); err != nil {
			serv.Log().WithField("error", err).Error("failed to generate backup for server")
		}
	}(adapter, s)

	c.Status(http.StatusAccepted)
}

// Deletes a local backup of a server.
func deleteServerBackup(c *gin.Context) {
	s := GetServer(c.Param("server"))

	b, _, err := backup.LocateLocal(c.Param("backup"))
	if err != nil {
		TrackedServerError(err, s).AbortWithServerError(c)
		return
	}

	if err := b.Remove(); err != nil {
		TrackedServerError(err, s).AbortWithServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
