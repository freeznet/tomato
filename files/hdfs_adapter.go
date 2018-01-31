package files

import (
	"errors"
	"net/url"
	"github.com/freeznet/tomato/config"
	"github.com/colinmarc/hdfs"
	"strings"
)

type hdfsAdapter struct {
	client *hdfs.Client
	rootpath string
}

func newHDFSAdapter(nn, user, rootpath string) *hdfsAdapter {
	g := &hdfsAdapter{}
	client, err := hdfs.NewForUser(nn, user)
	if err != nil {
		return nil
	}
	g.client = client
	g.rootpath = rootpath
	if !strings.HasSuffix(g.rootpath, "/") {
		g.rootpath = g.rootpath + "/"
	}
	return g
}

func (g *hdfsAdapter) createFile(filename string, data []byte, contentType string) error {
	file, err := g.client.Create(g.rootpath + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("createFile failed")
	}

	return nil
}

func (g *hdfsAdapter) deleteFile(filename string) error {
	return g.client.Remove(g.rootpath + filename)
}

func (g *hdfsAdapter) getFileData(filename string) ([]byte, error) {
	file, err := g.client.Open(g.rootpath + filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := []byte{}
	buf := make([]byte, 1024)
	for {
		n, _ := file.Read(buf)
		if n == 0 {
			break
		}
		data = append(data, buf[:n]...)
	}

	return data, nil
}

func (g *hdfsAdapter) getFileLocation(filename string) string {
	return config.TConfig.ServerURL + "/files/" + config.TConfig.AppID + "/" + url.QueryEscape(filename)
}

func (g *hdfsAdapter) getFileStream(filename string) (FileStream, error) {
	file, err := g.client.Open(g.rootpath + filename)
	if err != nil {
		return nil, err
	}
	return &hdfsFileStream{file: file}, nil
}

func (g *hdfsAdapter) getAdapterName() string {
	return "hdfsAdapter"
}

type hdfsFileStream struct {
	file *hdfs.FileReader
}

func (d *hdfsFileStream) Seek(offset int64, whence int) (ret int64, err error) {
	return d.file.Seek(offset, whence)
}

func (d *hdfsFileStream) Read(b []byte) (n int, err error) {
	return d.file.Read(b)
}

func (d *hdfsFileStream) Size() (bytes int64) {
	i := d.file.Stat()
	return i.Size()
}

func (d *hdfsFileStream) Close() (err error) {
	return d.file.Close()
}

