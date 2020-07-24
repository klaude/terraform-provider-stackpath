package stackpath

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-stackpath/stackpath/api/workload/workload_client/instances"
	"github.com/terraform-providers/terraform-provider-stackpath/stackpath/api/workload/workload_client/workloads"
	"github.com/terraform-providers/terraform-provider-stackpath/stackpath/api/workload/workload_models"
)

// annotation keys that should be ignored when diffing the state of a workload
var ignoredComputeWorkloadAnnotations = map[string]bool{
	"anycast.platform.stackpath.net/subnets": true,
}

func resourceComputeWorkload() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeWorkloadCreate,
		Read:   resourceComputeWorkloadRead,
		Update: resourceComputeWorkloadUpdate,
		Delete: resourceComputeWorkloadDelete,
		Importer: &schema.ResourceImporter{
			State: resourceComputeWorkloadImportState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: func(key, _, _ string, d *schema.ResourceData) bool {
					o, n := d.GetChange("annotations")
					oldData, newData := o.(map[string]interface{}), n.(map[string]interface{})
					for k, newVal := range newData {
						// check if it is an ignored annotation
						if ignoredComputeWorkloadAnnotations[k] {
							continue
						}
						// compare the previous value and see if it changed
						if oldVal, exists := oldData[k]; !exists || oldVal != newVal {
							return false
						}
					}

					for k, oldVal := range oldData {
						// check if it is an ignored annotation
						if ignoredComputeWorkloadAnnotations[k] {
							continue
						}
						// compare the previous value and see if it changed
						if newVal, exists := newData[k]; !exists || oldVal != newVal {
							return false
						}
					}

					return true
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"image_pull_credentials": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"docker_registry": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"server": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
									"password": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"email": {
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
									},
								},
							},
						},
					},
				},
			},
			"virtual_machine": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"container"},
				MaxItems:      1,
				Optional:      true,
				Elem:          resourceComputeWorkloadVirtualMachine(),
			},
			"container": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"virtual_machine"},
				Elem:          resourceComputeWorkloadContainer(),
			},
			"volume_claim": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resources": resourceComputeWorkloadResourcesSchema(),
					},
				},
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"min_replicas": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"max_replicas": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"scale_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metrics": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metric": {
													Type:     schema.TypeString,
													Required: true,
												},
												"average_utilization": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"average_value": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"deployment_scope": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "cityCode",
						},
						"selector": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem:     resourceComputeMatchExpressionSchema(),
						},
					},
				},
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     resourceComputeWorkloadInstance(),
			},
		},
	}
}

func resourceComputeWorkloadVolumeMountSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"slug": {
					Type:     schema.TypeString,
					Required: true,
				},
				"mount_path": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func resourceComputeWorkloadProbeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_get": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"path": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "/",
							},
							"port": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"scheme": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "http",
							},
							"http_headers": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"tcp_socket": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"port": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
				"initial_delay_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  0,
				},
				"timeout_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  10,
				},
				"period_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  60,
				},
				"success_threshold": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"failure_threshold": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}
}

func resourceComputeWorkloadResourcesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"requests": {
					Type:     schema.TypeMap,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func resourceComputeWorkloadPortSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"enable_implicit_network_policy": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"port": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"protocol": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "tcp",
				},
			},
		},
	}
}

func resourceComputeWorkloadCreate(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	// Create the workload
	resp, err := config.edgeCompute.Workloads.CreateWorkload(&workloads.CreateWorkloadParams{
		Context: context.Background(),
		StackID: config.StackID,
		Body: &workload_models.V1CreateWorkloadRequest{
			Workload: convertComputeWorkload(data),
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create compute workload: %v", NewStackPathError(err))
	}

	// Set the ID based on the workload created in the API
	data.SetId(resp.Payload.Workload.ID)

	lastInstancePhases := make(map[string]workload_models.Workloadv1InstanceInstancePhase, 0)

	wait := resource.StateChangeConf{
		Pending: []string{"starting"},
		Target: []string{"done"},
		Timeout: 5 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			status := "starting"
			allStarted := true
			workloadInstances := make([]workload_models.Workloadv1Instance, 0)

			// Get all workload instances in groups of 50
			pageSize := "50"
			var endCursor string
			for {
				params := &instances.GetWorkloadInstancesParams{
					StackID:          config.StackID,
					WorkloadID:       data.Id(),
					Context:          context.Background(),
					PageRequestFirst: &pageSize,
				}
				if endCursor != "" {
					params.PageRequestAfter = &endCursor
				}

				resp, err := config.edgeCompute.Instances.GetWorkloadInstances(params, nil)
				if err != nil {
					return nil, "", fmt.Errorf("failed to read compute workload instances: %v", NewStackPathError(err))
				}

				for _, result := range resp.Payload.Results {
					if result != nil {
						workloadInstances = append(workloadInstances, *result)
					}
				}

				// Continue paginating until we get all the results
				if !resp.Payload.PageInfo.HasNextPage {
					break
				}
				endCursor = resp.Payload.PageInfo.EndCursor
			}

			// If there aren't any workload instances then the workload was
			// created but instances aren't yet. Short circuit early to wait for
			// instances to show up.
			if len(workloadInstances) == 0 {
				return resp, status, nil
			}

			for _, instance := range workloadInstances {
				if phase, found := lastInstancePhases[instance.Name]; found {
					if phase != instance.Phase {
						log.Printf(
							"[INFO] Workload %s instance %s: %s -> %s",
							resp.Payload.Workload.Name,
							instance.Name,
							strings.ToLower(string(phase)),
							strings.ToLower(string(instance.Phase)),
						)
					}
				} else {
					log.Printf(
						"[INFO] Workload %s instance %s: %s",
						resp.Payload.Workload.Name,
						instance.Name,
						strings.ToLower(string(instance.Phase)),
					)
				}

				lastInstancePhases[instance.Name] = instance.Phase
			}

			for name, phase := range lastInstancePhases {
				if phase == workload_models.Workloadv1InstanceInstancePhaseSTARTING {
					allStarted = false
				}

				if phase == workload_models.Workloadv1InstanceInstancePhaseFAILED {
					log.Printf("[ERROR] Instance %s failed to start", name)
					return nil, "", fmt.Errorf(
						"workload %s instance %s failed to start",
						resp.Payload.Workload.Name,
						name,
					)
				}
			}

			if allStarted {
				log.Printf("[INFO] All workload %s instances have started", resp.Payload.Workload.Name)
				status = "done"
			}

			return resp, status, err
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	return resourceComputeWorkloadRead(data, meta)
}

func resourceComputeWorkloadUpdate(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	_, err := config.edgeCompute.Workloads.UpdateWorkload(&workloads.UpdateWorkloadParams{
		Context:    context.Background(),
		StackID:    config.StackID,
		WorkloadID: data.Id(),
		Body: &workload_models.V1UpdateWorkloadRequest{
			Workload: convertComputeWorkload(data),
		},
	}, nil)
	if c, ok := err.(interface{ Code() int }); ok && c.Code() == http.StatusNotFound {
		// Clear out the ID in terraform if the
		// resource no longer exists in the API
		data.SetId("")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to update compute workload: %v", NewStackPathError(err))
	}
	return resourceComputeWorkloadRead(data, meta)
}

func resourceComputeWorkloadRead(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	resp, err := config.edgeCompute.Workloads.GetWorkload(&workloads.GetWorkloadParams{
		Context:    context.Background(),
		StackID:    config.StackID,
		WorkloadID: data.Id(),
	}, nil)
	if c, ok := err.(interface{ Code() int }); ok && c.Code() == http.StatusNotFound {
		// Clear out the ID in terraform if the
		// resource no longer exists in the API
		data.SetId("")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read compute workload: %v", NewStackPathError(err))
	}

	if err := flattenComputeWorkload(data, resp.Payload.Workload); err != nil {
		return err
	}

	return resourceComputeWorkloadReadInstances(data, meta)
}

func resourceComputeWorkloadReadInstances(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	pageSize := "50"
	// variable to keep track of our location through pagination
	var endCursor string
	var terraformInstances []interface{}
	for {
		params := &instances.GetWorkloadInstancesParams{
			StackID:          config.StackID,
			WorkloadID:       data.Id(),
			Context:          context.Background(),
			PageRequestFirst: &pageSize,
		}
		if endCursor != "" {
			params.PageRequestAfter = &endCursor
		}
		resp, err := config.edgeCompute.Instances.GetWorkloadInstances(params, nil)
		if err != nil {
			return fmt.Errorf("failed to read compute workload instances: %v", NewStackPathError(err))
		}
		for _, result := range resp.Payload.Results {
			terraformInstances = append(terraformInstances, flattenComputeWorkloadInstance(result))
		}
		// Continue paginating until we get all the results
		if !resp.Payload.PageInfo.HasNextPage {
			break
		}
		endCursor = resp.Payload.PageInfo.EndCursor
	}

	if err := data.Set("instances", terraformInstances); err != nil {
		return fmt.Errorf("error setting instances: %v", err)
	}

	return nil
}

func resourceComputeWorkloadDelete(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	_, err := config.edgeCompute.Workloads.DeleteWorkload(&workloads.DeleteWorkloadParams{
		Context:    context.Background(),
		StackID:    config.StackID,
		WorkloadID: data.Id(),
	}, nil)
	if c, ok := err.(interface{ Code() int }); ok && c.Code() == http.StatusNotFound {
		// Clear out the ID in terraform if the
		// resource no longer exists in the API
		data.SetId("")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to delete compute workload: %v", NewStackPathError(err))
	}
	return nil
}

func resourceComputeWorkloadImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// We expect that to import a resource, the user will pass in
	// the full UUID of the workload they're attempting to import.
	return []*schema.ResourceData{d}, nil
}
