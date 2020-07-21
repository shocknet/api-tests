package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os/exec"

	"github.com/boufni95/goutils"
)

func generateDeviceID(r *rand.Rand) (string, error) {
	p1 := make([]byte, 4)
	p2 := make([]byte, 2)
	p3 := make([]byte, 1)
	p4 := make([]byte, 1)
	p5 := make([]byte, 6)
	_, err := r.Read(p1)
	if err != nil {
		return "", err
	}
	_, err = r.Read(p2)
	if err != nil {
		return "", err
	}
	_, err = r.Read(p3)
	if err != nil {
		return "", err
	}
	_, err = r.Read(p4)
	if err != nil {
		return "", err
	}
	_, err = r.Read(p5)
	if err != nil {
		return "", err
	}
	s1 := hex.EncodeToString(p1)
	s2 := hex.EncodeToString(p2)
	s3 := hex.EncodeToString(p3)
	s4 := hex.EncodeToString(p4)
	s5 := hex.EncodeToString(p5)
	res := s1 + "-" + s2 + "-40" + s3 + "-95" + s4 + "-" + s5
	return res, nil
}

func GenerateAuthPair(base string, r *rand.Rand) (NewUserReq, error) {
	req := NewUserReq{}
	alias := make([]byte, 3)
	pass := make([]byte, 3)
	_, err := r.Read(alias)
	if err != nil {
		return req, err
	}
	_, err = r.Read(pass)
	if err != nil {
		return req, err
	}
	aliasS := hex.EncodeToString(alias)
	passS := hex.EncodeToString(pass)
	req.Alias = base + aliasS
	req.Pass = passS
	return req, nil

}
func execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmdOutput := &bytes.Buffer{}
	cmdErr := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdErr
	err := cmd.Run()
	//fmt.Println(string(cmdErr.Bytes()))
	return cmdOutput.Bytes(), err
}

func executeAsync(name string, args ...string) {
	goutils.Log("starting async command " + name + " " + args[0])
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	go copyOutput(stdout)
	go copyOutput(stderr)
	cmd.Wait()
	goutils.Log("async command stopped")
}

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
func checkResError(data []byte) {
	resErr := ErrorsWrapper{}
	err := goutils.JsonBytesToType(data, &resErr)
	if err != nil {
		goutils.PrintError(err)
		return
	}
	if resErr.ErrorMessage != "" {
		goutils.PrintError(errors.New(resErr.ErrorMessage))
		return
	}
	if resErr.Err != "" {
		goutils.PrintError(errors.New(resErr.Err))
		return
	}
	goutils.Log(string(data))
}

func encryptTypeToString(info TestInfo, data interface{}) (string, error) {
	dataBytes, err := goutils.ToJsonBytes(data)
	if err != nil {
		return "", err
	}
	encBytes, err := info.Encrypt(dataBytes)
	if err != nil {
		return "", err
	}
	return goutils.ToJsonString(encBytes)
}
