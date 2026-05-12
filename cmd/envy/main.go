package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func usage() {
	fmt.Print(`envy — Environment variable manager

Usage: envy <command> [options]

Commands:
  list              List all environment variables
  get <key>         Get value for key
  set <key> <val>   Set key=value (persist to ~/.env)
  unset <key>       Remove key from ~/.env
  load <file>       Load .env file into current shell
  check             Check which .env vars are set
  encrypt           Encrypt values (placeholder)
  decrypt           Decrypt values (placeholder)

Options:
  -f string         Env file path (default ~/.env)
  -e                Include empty values
  -s                Sort output

Examples:
  envy list
  envy get DATABASE_URL
  envy set API_KEY sk-abc123
  envy load .env.local
  envy check
`)
}

func main() {
	var (
		envFile = flag.String("f", filepath.Join(os.Getenv("HOME"), ".env"), "Env file path")
		showEmpty = flag.Bool("e", false, "Include empty values")
		sortKeys = flag.Bool("s", false, "Sort output")
	)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	switch cmd {
	case "list", "ls":
		runList(*showEmpty, *sortKeys)
	case "get":
		if len(args) < 1 {
			fmt.Println("Usage: envy get <key>")
			os.Exit(1)
		}
		runGet(args[0])
	case "set":
		if len(args) < 2 {
			fmt.Println("Usage: envy set <key> <value>")
			os.Exit(1)
		}
		runSet(args[0], args[1], *envFile)
	case "unset", "rm":
		if len(args) < 1 {
			fmt.Println("Usage: envy unset <key>")
			os.Exit(1)
		}
		runUnset(args[0], *envFile)
	case "load":
		if len(args) < 1 {
			fmt.Println("Usage: envy load <file>")
			os.Exit(1)
		}
		runLoad(args[0])
	case "check":
		runCheck(*envFile)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func runList(showEmpty, sortKeys bool) {
	env := os.Environ()
	if sortKeys {
		sort.Strings(env)
	}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) > 1 {
			val = parts[1]
		}
		if !showEmpty && val == "" {
			continue
		}
		// Mask secrets
		if isSecret(key) && len(val) > 8 {
			val = val[:4] + "****" + val[len(val)-4:]
		}
		fmt.Printf("%s=%s\n", key, val)
	}
}

func runGet(key string) {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "Key not set: %s\n", key)
		os.Exit(1)
	}
	fmt.Println(val)
}

func runSet(key, value, envFile string) {
	envMap := loadEnvFile(envFile)
	envMap[key] = value
	if err := saveEnvFile(envFile, envMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Set %s in %s\n", key, envFile)
}

func runUnset(key, envFile string) {
	envMap := loadEnvFile(envFile)
	if _, exists := envMap[key]; !exists {
		fmt.Fprintf(os.Stderr, "Key not found in %s: %s\n", envFile, key)
		os.Exit(1)
	}
	delete(envMap, key)
	if err := saveEnvFile(envFile, envMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Removed %s from %s\n", key, envFile)
}

func runLoad(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			fmt.Printf("export %s=%q\n", parts[0], parts[1])
		}
	}
}

func runCheck(envFile string) {
	envMap := loadEnvFile(envFile)
	if len(envMap) == 0 {
		fmt.Printf("No vars in %s\n", envFile)
		return
	}

	var keys []string
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("Checking %d vars from %s:\n\n", len(keys), envFile)
	for _, k := range keys {
		val := os.Getenv(k)
		if val != "" {
			fmt.Printf("✓ %s (set)\n", k)
		} else {
			fmt.Printf("✗ %s (missing)\n", k)
		}
	}
}

func loadEnvFile(path string) map[string]string {
	result := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func saveEnvFile(path string, envMap map[string]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	var keys []string
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, k := range keys {
		fmt.Fprintf(f, "%s=%s\n", k, envMap[k])
	}
	return nil
}

func isSecret(key string) bool {
	lower := strings.ToLower(key)
	secrets := []string{"key", "token", "secret", "password", "pass", "auth", "credential"}
	for _, s := range secrets {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
