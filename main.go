package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type GPUData struct {
	ID          string `json:"gpu_id"`
	Temperature string `json:"temperature"`
	MemoryUsage string `json:"memory_usage,omitempty"`
	GPUUtil     string `json:"gpu_util"`
	Power       string `json:"power,omitempty"`
}

func getNvidiaGPUData() ([]GPUData, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,temperature.gpu,memory.used,memory.total,utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseNvidiaGPUData(output), nil
}

func parseNvidiaGPUData(output []byte) []GPUData {
	lines := strings.Split(string(output), "\n")
	var gpuData []GPUData
	for _, line := range lines {
		if line != "" {
			fields := strings.Split(line, ", ")
			if len(fields) >= 5 {
				gpu := GPUData{
					ID:          fields[0],
					Temperature: fields[1] + "°C",
					MemoryUsage: fields[2] + " / " + fields[3] + " MiB",
					GPUUtil:     fields[4] + "%",
				}
				gpuData = append(gpuData, gpu)
			}
		}
	}
	return gpuData
}

func getAMDGPUData() ([]GPUData, error) {
	cmd := exec.Command("rocm-smi", "--showtemp", "--showuse", "--json", "--showperflevel", "--showpower")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var data map[string]map[string]string
	err = json.Unmarshal(output, &data)
	if err != nil {
		return nil, err
	}
	var gpuData []GPUData
	for card, info := range data {
		gpu := GPUData{
			ID:          card,
			Temperature: info["Temperature (Sensor edge) (C)"],
			GPUUtil:     info["GPU use (%)"],
			Power:       info["Average Graphics Package Power (W)"],
		}
		gpuData = append(gpuData, gpu)
	}
	return gpuData, nil
}

func detectGPUBrand() (string, error) {
	if err := exec.Command("nvidia-smi").Run(); err == nil {
		return "nvidia", nil
	}
	if err := exec.Command("rocm-smi").Run(); err == nil {
		return "amd", nil
	}
	return "", fmt.Errorf("unable to detect GPU brand")
}

func gpuHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	brand, err := detectGPUBrand()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var gpuData []GPUData
	if brand == "nvidia" {
		gpuData, err = getNvidiaGPUData()
	} else {
		gpuData, err = getAMDGPUData()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(gpuData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	http.HandleFunc("/gpustats", gpuHandler)
	http.ListenAndServe(":8080", nil)
}
