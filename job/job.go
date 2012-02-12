package job

import (
	"gonzbee/config"
	"gonzbee/nntp"
	"gonzbee/nzb"
	"gonzbee/yenc"
	"os"
	"path"
	"path/filepath"
)

//The Job struct holds all the information needed in order to download
//a posting from an NZB file
type Job struct {
	Name string
	Nzb  *nzb.Nzb
}

//FromFile creates a download job from a NZB file
func FromFile(filepath string) (*Job, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	nzbFile, err := nzb.Parse(file)
	if err != nil {
		return nil, err
	}
	j := &Job{Name: path.Base(filepath), Nzb: nzbFile}
	return j, nil
}

//Start will execute a job on the given NNTP connection
func (j *Job) Start(nntpConn *nntp.Conn) error {
	path := config.C.GetIncompleteDir()
	jobDir := filepath.Join(path, j.Name)
	os.Mkdir(jobDir, 0777)
	for _, file := range j.Nzb.File {
		nntpConn.SwitchGroup(file.Groups[0])
		for _, seg := range file.Segments {
			contents, err := nntpConn.GetMessageReader(seg.MsgId)
			if err != nil {
				continue
			}
			part, _ := yenc.NewPart(contents)
			file, _ := os.OpenFile(filepath.Join(jobDir, part.Name), os.O_WRONLY|os.O_CREATE, 0644)
			file.Seek(part.Begin, os.SEEK_SET)
			part.Decode(file)
			file.Close()
			contents.Close()
		}
	}
	return nil
}
