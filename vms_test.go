package main

import (
	"github.com/esxcloud/bosh-esxcloud-cpi/cpi"
	. "github.com/esxcloud/bosh-esxcloud-cpi/mocks"
	ec "github.com/esxcloud/esxcloud-go-sdk/esxcloud"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("VMs", func() {
	var (
		server *httptest.Server
		ctx    *cpi.Context
		projID string
	)

	BeforeEach(func() {
		server = NewMockServer()

		Activate(true)
		httpClient := &http.Client{Transport: DefaultMockTransport}
		ctx = &cpi.Context{
			Client: ec.NewTestClient(server.URL, httpClient),
			Config: &cpi.Config{
				ESXCloud: &cpi.ESXCloudConfig{
					APIFE:     server.URL,
					ProjectID: "fake-project-id",
				},
			},
		}

		projID = ctx.Config.ESXCloud.ProjectID
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CreateVM", func() {
		It("should return ID of created VM", func() {
			createTask := &ec.Task{Operation: "CREATE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "CREATE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/v1/projects/"+projID+"/vms",
				CreateResponder(200, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/v1/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"unused-agent-id", "fake-stemcell-id", map[string]interface{}{"flavor": "fake-flavor"}}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(Equal(completedTask.Entity.ID))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when server returns error", func() {
			createTask := &ec.Task{Operation: "CREATE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "CREATE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"POST",
				server.URL+"/v1/projects/"+projID+"/vms",
				CreateResponder(500, ToJson(createTask)))
			RegisterResponder(
				"GET",
				server.URL+"/v1/tasks/"+createTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"unused-agent-id", "fake-stemcell-id", map[string]interface{}{"flavor": "fake-flavor"}}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when cloud_properties has bad property type", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"unused-agent-id", "fake-stemcell-id", map[string]interface{}{"flavor": 123}}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when cloud_properties has no properties", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"unused-agent-id", "fake-stemcell-id", map[string]interface{}{}}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when cloud_properties is missing", func() {
			actions := map[string]cpi.ActionFn{
				"create_vm": CreateVM,
			}
			args := []interface{}{"unused-agent-id", "fake-stemcell-id"}
			res, err := GetResponse(dispatch(ctx, actions, "create_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("DeleteVM", func() {
		It("should return nothing when successful", func() {
			deleteTask := &ec.Task{Operation: "DELETE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "DELETE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/v1/vms/"+deleteTask.Entity.ID+"?force=true",
				CreateResponder(200, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+"/v1/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when VM not found", func() {
			deleteTask := &ec.Task{Operation: "DELETE_VM", State: "QUEUED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}
			completedTask := &ec.Task{Operation: "DELETE_VM", State: "COMPLETED", ID: "fake-task-id", Entity: ec.Entity{ID: "fake-vm-id"}}

			RegisterResponder(
				"DELETE",
				server.URL+"/v1/vms/"+deleteTask.Entity.ID+"?force=true",
				CreateResponder(404, ToJson(deleteTask)))
			RegisterResponder(
				"GET",
				server.URL+"/v1/tasks/"+deleteTask.ID,
				CreateResponder(200, ToJson(completedTask)))

			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"delete_vm": DeleteVM,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "delete_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
	Describe("HasVM", func() {
		It("should return true when VM is found", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/v1/vms/"+vm.ID,
				CreateResponder(200, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(Equal(true))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return false when VM not found", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/v1/vms/"+vm.ID,
				CreateResponder(404, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(Equal(false))
			Expect(res.Error).Should(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when server returns error", func() {
			vm := &ec.VM{ID: "fake-vm-id"}
			RegisterResponder(
				"GET",
				server.URL+"/v1/vms/"+vm.ID,
				CreateResponder(500, ToJson(vm)))

			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{"fake-vm-id"}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when given no arguments", func() {
			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should return an error when given an invalid argument", func() {
			actions := map[string]cpi.ActionFn{
				"has_vm": HasVM,
			}
			args := []interface{}{5}
			res, err := GetResponse(dispatch(ctx, actions, "has_vm", args))

			Expect(res.Result).Should(BeNil())
			Expect(res.Error).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})