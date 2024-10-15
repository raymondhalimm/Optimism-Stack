package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Helper function to run a command
func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

// Function to change the working directory
func changeDir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("Failed to change directory to %s: %v", dir, err)
	}
	fmt.Printf("Changed directory to: %s\n", dir)
}

// Function to fill environment variables in .envrc (L1 URL and Kind)
func fillOutEnvVariables(filePath, l1RpcUrl, l1RpcKind string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Read the file into memory
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Replace placeholders with actual values
		if strings.HasPrefix(line, "L1_RPC_URL") {
			line = fmt.Sprintf("L1_RPC_URL=%s", l1RpcUrl)
		} else if strings.HasPrefix(line, "L1_RPC_KIND") {
			line = fmt.Sprintf("L1_RPC_KIND=%s", l1RpcKind)
		}

		lines = append(lines, line)
	}

	// Check for any scanning errors
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Write the updated content back to the original file
	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to open file for writing: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}
	writer.Flush()

	fmt.Printf("Updated environment variables in %s\n", filePath)
}

// Function to run wallets.sh and capture its output
func runCommandAndCaptureOutput(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	return string(output)
}

// Function to fill environment variables in .envrc (Wallets addresses and keys)
func replacePlaceholdersInEnvrc(filename, output string) {
	// Open the envrc file for reading
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Prepare to read the file and replace placeholders
	scanner := bufio.NewScanner(file)
	var fileContent []string

	// Parse the wallets.sh output
	lines := strings.Split(output, "\n")
	var replacements = map[string]string{}

	// Parse the script output to extract addresses and private keys
	for _, line := range lines {
		if strings.HasPrefix(line, "export GS_") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				replacements[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Read the envrc file line by line and replace placeholders
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line contains an export statement to replace
		for key, value := range replacements {
			if strings.HasPrefix(line, key) {
				line = fmt.Sprintf("%s=%s", key, value)
			}
		}

		fileContent = append(fileContent, line)
	}

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Write the updated content back to the envrc file
	err = os.WriteFile(filename, []byte(strings.Join(fileContent, "\n")), 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	fmt.Printf("Updated %s successfully.\n", filename)
}

// Function to open a new terminal and start op
func startInNewTerminal(script string) {
	// AppleScript to open a new terminal and run the op-geth command

	// Execute the AppleScript using osascript
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to open terminal : %v", err)
	}
}

func main() {
	// Clone the Optimism repository
	fmt.Println("Cloning Optimism repository...")
	runCommand("git", "clone", "https://github.com/ethereum-optimism/optimism.git")

	// Check out the branch
	changeDir("~/optimism")
	runCommand("git", "checkout", "tutorials/chain")

	// Install dependencies and build
	fmt.Println("Installing dependencies...")
	runCommand("pnpm", "install")
	fmt.Println("Building components...")
	runCommand("make", "op-node", "op-batcher", "op-proposer")
	runCommand("pnpm", "build")

	// Clone the op-geth repository and build geth
	fmt.Println("Cloning op-geth repository...")
	changeDir("~")
	runCommand("git", "clone", "https://github.com/ethereum-optimism/op-geth.git")

	// Build op-geth
	changeDir("~/op-geth")
	fmt.Println("Building geth...")
	runCommand("make", "geth")

	// Duplicate the .envrc.example file
	changeDir("~/optimism")
	fmt.Println("Copying .envrc.example to .envrc...")
	runCommand("cp", ".envrc.example", ".envrc")

	// Fill out the environment variables in .envrc
	// L1_RPC_URL & L1_RPC_KIND respectively
	fmt.Println("Filling out environment variables in .envrc...")
	fillOutEnvVariables("~/optimism/.envrc", "https://eth-sepolia.g.alchemy.com/v2/aw63uoS50a3OYBabEf5H62HfJnaozM3p", "alchemy")

	// Generate new addresses by running wallets.sh
	changeDir("~/optimism")
	fmt.Println("Generating new addresses...")
	output := runCommandAndCaptureOutput("./packages/contracts-bedrock/scripts/getting-started/wallets.sh")

	// Replace the placeholders of address and key in the .envrc file
	fmt.Println("Replacing placeholders in the .envrc file...")
	replacePlaceholdersInEnvrc("~/optimism/.envrc", output)

	// Load environment variables with direnv
	changeDir("~/optimism")
	fmt.Println("Loading environment variables with direnv...")
	runCommand("direnv", "allow")

	// Move into the contracts-bedrock package and generate the config file
	changeDir("~/optimism/packages/contracts-bedrock")
	fmt.Println("Generating getting-started.json configuration file...")
	runCommand("./scripts/getting-started/config.sh")

	// Deploy the L1 contracts using forge
	fmt.Println("Deploying L1 contracts...")
	runCommand("forge", "script", "scripts/Deploy.s.sol:Deploy", "--private-key", os.Getenv("GS_ADMIN_PRIVATE_KEY"), "--broadcast", "--rpc-url", os.Getenv("L1_RPC_URL"), "--slow")

	// Generate the L2 Config Files (Create genesis.json and rollup.json files)
	changeDir("~/optimism/op-node")
	fmt.Println("Creating genesis and rollup files...")
	runCommand("go", "run", "cmd/main.go", "genesis", "l2",
		"--deploy-config", "../packages/contracts-bedrock/deploy-config/getting-started.json",
		"--l1-deployments", "../packages/contracts-bedrock/deployments/getting-started/.deploy",
		"--outfile.l2", "genesis.json",
		"--outfile.rollup", "rollup.json",
		"--l1-rpc", os.Getenv("L1_RPC_URL"))

	// Create a JSON Web Token (JWT) for authentication
	fmt.Println("Creating JWT token...")
	runCommand("openssl", "rand", "-hex", "32", ">", "jwt.txt")

	// Copy genesis.json and jwt.txt into the op-geth directory
	fmt.Println("Copying genesis.json and jwt.txt into op-geth...")
	runCommand("cp", "genesis.json", "~/op-geth")
	runCommand("cp", "jwt.txt", "~/op-geth")

	// Create a data directory folder
	changeDir("~/op-geth")
	fmt.Println("Creating data directory for op-geth...")
	runCommand("mkdir", "datadir")

	// Initialize op-geth with genesis.json
	// Input "--state.scheme=hash" from the original command to be able to run
	fmt.Println("Initializing op-geth with genesis.json...")
	runCommand("build/bin/geth", "init", "--datadir=datadir", "--state.scheme=hash", "genesis.json")

	// Open a new terminal and run the op-geth
	startInNewTerminal(`
	tell application "Terminal"
		do script "cd ~/op-geth && ./build/bin/geth \
			--datadir ./datadir \
			--http \
			--http.corsdomain=* \
			--http.vhosts=* \
			--http.addr=0.0.0.0 \
			--http.api=web3,debug,eth,txpool,net,engine \
			--ws \
			--ws.addr=0.0.0.0 \
			--ws.port=8546 \
			--ws.origins=* \
			--ws.api=debug,eth,txpool,net,engine \
			--syncmode=full \
			--gcmode=archive \
			--nodiscover \
			--maxpeers=0 \
			--networkid=42069 \
			--authrpc.vhosts=* \
			--authrpc.addr=0.0.0.0 \
			--authrpc.port=8551 \
			--authrpc.jwtsecret=./jwt.txt \
			--rollup.disabletxpoolgossip=true"
		activate
	end tell`)
	fmt.Println("op-geth started in a new terminal.")

	// Open a new terminal and run the op-node
	startInNewTerminal(`
	tell application "Terminal"
		do script "cd ~/optimism/op-node && ./bin/op-node \
			--l2=http://localhost:8551 \
			--l2.jwt-secret=./jwt.txt \
			--sequencer.enabled \
			--sequencer.l1-confs=5 \
			--verifier.l1-confs=4 \
			--rollup.config=./rollup.json \
			--rpc.addr=0.0.0.0 \
			--p2p.disable \
			--rpc.enable-admin \
			--p2p.sequencer.key=$GS_SEQUENCER_PRIVATE_KEY \
			--l1=$L1_RPC_URL \
			--l1.rpckind=$L1_RPC_KIND"
		activate
	end tell`)
	fmt.Println("op-node started in a new terminal.")

	// Open a new terminal and run the op-batcher
	startInNewTerminal(`
	tell application "Terminal"
		do script "cd ~/optimism/op-batcher && ./bin/op-batcher \
			--l2-eth-rpc=http://localhost:8545 \
			--rollup-rpc=http://localhost:9545 \
			--poll-interval=1s \
			--sub-safety-margin=6 \
			--num-confirmations=1 \
			--safe-abort-nonce-too-low-count=3 \
			--resubmission-timeout=30s \
			--rpc.addr=0.0.0.0 \
			--rpc.port=8548 \
			--rpc.enable-admin \
			--max-channel-duration=25 \
			--l1-eth-rpc=$L1_RPC_URL \
			--private-key=$GS_BATCHER_PRIVATE_KEY"
		activate
	end tell`)
	fmt.Println("op-batcher started in a new terminal.")

	// Open a new terminal and run the op-proposer
	startInNewTerminal(`
	tell application "Terminal"
		do script "cd ~/optimism/op-proposer && ./bin/op-proposer \
			--poll-interval=12s \
			--rpc.port=8560 \
			--rollup-rpc=http://localhost:9545 \
			--l2oo-address=$(cat ../packages/contracts-bedrock/deployments/getting-started/.deploy | jq -r .L2OutputOracleProxy) \
			--private-key=$GS_PROPOSER_PRIVATE_KEY \
			--l1-eth-rpc=$L1_RPC_URL"
		activate
	end tell`)
	fmt.Println("op-proposer started in a new terminal.")

	fmt.Println("Automation completed successfully!")
}
