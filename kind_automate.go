package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// Helper function to run shell commands
func runCommand(cmd string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("error: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Function to create the kind cluster if it doesn't exist
func createKindCluster(clusterName string) error {
	fmt.Println("Creating kind cluster...")

	// Check if the cluster already exists
	output, err := runCommand("kind", "get", "clusters")
	if err != nil {
		return err
	}

	if !bytes.Contains([]byte(output), []byte(clusterName)) {
		// Create a new kind cluster
		_, err = runCommand("kind", "create", "cluster", "--name", clusterName)
		if err != nil {
			return fmt.Errorf("failed to create kind cluster: %v", err)
		}
	} else {
		fmt.Printf("Kind cluster '%s' already exists.\n", clusterName)
	}
	return nil
}

// Function to install Helm (if needed) and add the repository
func setupHelmRepo(repoName, repoURL string) error {

	// Add the Helm repository from Artifact Hub
	_, err := runCommand("helm", "repo", "add", repoName, repoURL)
	if err != nil {
		return fmt.Errorf("failed to add Helm repository: %v", err)
	}

	// Update Helm repositories
	_, err = runCommand("helm", "repo", "update")
	if err != nil {
		return fmt.Errorf("failed to update Helm repositories: %v", err)
	}

	return nil
}

// Function to install the OP Stack Helm chart
func installHelmChart(chartName, releaseName, namespace string) error {
	fmt.Println("Installing the OP Stack Helm chart...")
	_, err := runCommand("helm", "install", releaseName, chartName, "--namespace", namespace, "--create-namespace")
	if err != nil {
		return fmt.Errorf("failed to install Helm chart: %v", err)
	}
	return nil
}

// Main automation function
func automateDeployment() error {

	// Step 1: Create the kind cluster
	clusterName := "op-stack-cluster"
	err := createKindCluster(clusterName)
	if err != nil {
		return fmt.Errorf("error creating cluster: %v", err)
	}

	// Step 2: Wait for the cluster to be ready
	fmt.Println("Waiting for cluster to be ready...")
	time.Sleep(30 * time.Second)

	// Step 3: Set up the Helm repository (Add and Update)
	repoName := "op-chain-charts"
	repoURL := "https://geo-web-project.github.io/op-chain-charts/"
	err = setupHelmRepo(repoName, repoURL)
	if err != nil {
		return fmt.Errorf("error setting up Helm repo: %v", err)
	}

	// Step 4: Install the OP Stack Helm chart
	// Install op-geth helm chart
	chartName := fmt.Sprintf("%s/op-geth", repoName)
	namespace := "op-namespace"
	err = installHelmChart(chartName, "my-op-geth", namespace)
	if err != nil {
		return fmt.Errorf("error installing Helm chart: %v", err)
	}

	// Install op-node helm chart
	chartName = fmt.Sprintf("%s/op-node", repoName)
	err = installHelmChart(chartName, "my-op-node", namespace)
	if err != nil {
		return fmt.Errorf("error installing Helm chart: %v", err)
	}

	// Install op-batcher helm chart
	chartName = fmt.Sprintf("%s/op-batcher", repoName)
	err = installHelmChart(chartName, "my-op-batcher", namespace)
	if err != nil {
		return fmt.Errorf("error installing Helm chart: %v", err)
	}

	// Install op-proposer helm chart
	chartName = fmt.Sprintf("%s/op-proposer", repoName)
	err = installHelmChart(chartName, "my-op-proposer", namespace)
	if err != nil {
		return fmt.Errorf("error installing Helm chart: %v", err)
	}

	fmt.Println("OP stack deployed successfully in kind cluster!")
	return nil
}

func main() {
	err := automateDeployment()
	if err != nil {
		log.Fatalf("Automation failed: %v", err)
	}
}
