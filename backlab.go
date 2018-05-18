// Package backlab makes backing up your GitLab instance easy.
package backlab

import (
	"os"
	"path/filepath"

	"github.com/dchest/uniuri"
	"gopkg.in/kothar/go-backblaze.v0"
	"os/exec"
	"io/ioutil"
)

// Credentials is a type alias for backblaze.Credentials.
type Credentials backblaze.Credentials

// Config is a parameter used in Init for configuring backlab.
type Config struct {
	Credentials
	BucketName string
	// PreserveFor defines how long should the backups be kept in a bucket and locally. Doesn't delete any by default.
	PreserveFor int
	BackupPath string
}

// Backlab is a main struct.
type Backlab struct {
	Config

	b2 *backblaze.B2
}

// New configures connection with Backblaze and saves configuration to Backlab instance.
func New(config Config) (*Backlab, error) {
	b := &Backlab{
		Config: config,
	}

	b2, err := backblaze.NewB2(backblaze.Credentials{
		AccountID:      b.AccountID,
		ApplicationKey: b.ApplicationKey,
	})
	if err != nil {
		panic(err)
	}

	b.b2 = b2

	return b, nil
}

// Backup backups a file to Backblaze.
func (b *Backlab) BackupArchive(archivePath string) error {
	var (
		bucket *backblaze.Bucket
		err    error
	)
	if &b.BucketName == nil {
		randomString := uniuri.New()
		bucket, err = b.b2.CreateBucket("backlab-gitlab-backups-"+randomString, backblaze.AllPrivate)
	} else {
		bucket, err = b.b2.Bucket(b.BucketName)
	}
	if err != nil {
		return err
	}

	reader, _ := os.Open(archivePath)
	name := filepath.Base(archivePath)
	metadata := make(map[string]string)

	_, err = bucket.UploadFile(name, metadata, reader)
	if err != nil {
		return err
	}

	// scan bucket for backups older than specified and delete them

	return nil
}

func (b *Backlab) CreateBackup() error {
	cmd := exec.Command("gitlab-rake", "gitlab:backup:create")
	err := cmd.Run()
	return err
}

func (b *Backlab) RemoveOldLocalBackups() error {
	files, _ := ioutil.ReadDir(b.BackupPath)

	// TODO remove files older than now - PreserveFor
}
