package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esxcloud/bosh-esxcloud-cpi/cmd"
	"github.com/esxcloud/bosh-esxcloud-cpi/cpi"
	"io/ioutil"
	"os"
	p "path"
)

func createAgentEnv(ctx *cpi.Context, agentID, vmID, vmName string, networks, env map[string]interface{}) (res *cpi.AgentEnv) {
	res = &cpi.AgentEnv{
		AgentID:  agentID,
		VM:       cpi.VMSpec{Name: vmName, ID: vmID},
		Networks: networks,
		Env:      env,
		Mbus:     ctx.Config.Agent.Mbus,
		NTP:      ctx.Config.Agent.NTP,
		Disks:    map[string]interface{}{"ephemeral": "1"},
		Blobstore: cpi.BlobstoreSpec{
			Provider: ctx.Config.Agent.Blobstore.Provider,
			Options:  ctx.Config.Agent.Blobstore.Options,
		},
	}
	return
}

func getAgentEnvMetadata(vmID string) (res *cpi.AgentEnv, err error) {
	// TODO: replace temp file with VM metadata API
	file := p.Join(os.TempDir(), vmID)
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	res = &cpi.AgentEnv{}
	err = json.Unmarshal(buf, res)
	return
}

func putAgentEnvMetadata(vmID string, env *cpi.AgentEnv) (err error) {
	buf, err := json.Marshal(env)
	if err != nil {
		return
	}
	// TODO: replace temp file with VM metadata API
	file := p.Join(os.TempDir(), vmID)
	err = ioutil.WriteFile(file, buf, 0777)
	return
}

func createEnvISO(env *cpi.AgentEnv, runner cmd.Runner) (path string, err error) {
	json, err := json.Marshal(env)
	if err != nil {
		return
	}
	envDir, err := ioutil.TempDir("", "agent-iso-dir")
	if err != nil {
		return
	}
	// Name of the environment JSON file should be "env" to fit ISO 9660 8.3 filename scheme
	envFile, err := os.Create(p.Join(envDir, "env"))
	if err != nil {
		return
	}
	_, err = envFile.Write(json)
	if err != nil {
		return
	}
	err = envFile.Close()
	if err != nil {
		return
	}

	envISO, err := ioutil.TempFile("", "agent-env-iso")
	if err != nil {
		return
	}
	envISO.Close()
	output, err := runner.Run("genisoimage", "-o", envISO.Name(), envFile.Name())
	if err != nil {
		out := string(output[:])
		return "", errors.New(fmt.Sprintf("Failed to generate ISO for agent settings: %v\n%s", err, out))
	}
	// Cleanup temp dir but ignore the error. Failure to delete a temp file is not
	// worth worrying about.
	_ = os.RemoveAll(envDir)
	return envISO.Name(), nil
}
