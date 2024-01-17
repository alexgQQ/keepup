package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
)

// startInstance starts a stopped Google Compute Engine instance
func startInst(w io.Writer, projectID, zone, instName string) error {
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer instancesClient.Close()

	req := &computepb.StartInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instName,
	}

	op, err := instancesClient.Start(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to start instance: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %w", err)
	}

	return nil
}

// checkInstance checks if an instance is stopped or not.
func isInstStopped(w io.Writer, projectID, zone, instName string) (bool, error) {
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return false, fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer instancesClient.Close()

	getInstanceReq := &computepb.GetInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instName,
	}

	instance, err := instancesClient.Get(ctx, getInstanceReq)
	if err != nil {
		return false, fmt.Errorf("unable to get instance: %w", err)
	}

	return instance.GetStatus() == "TERMINATED", nil
}

func main() {

	if len(os.Args) != 4 {
		log.Fatal(fmt.Errorf("This requires 3 positional arguments: instance_name project_id gcp_zone"))
	}
	name := os.Args[1]
	project := os.Args[2]
	zone := os.Args[3]

	log.Printf("Checking instance %s status...\n", name)
	stopped, err := isInstStopped(os.Stdout, project, zone, name)
	if err != nil {
		log.Fatal(err)
	}

	if stopped {
		log.Printf("Instance %s is stopped, restarting it...\n", name)
		err = startInst(os.Stdout, project, zone, name)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("All Done!")
}
