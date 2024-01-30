package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,temperature.gpu,memory.used,memory.total,utilization.gpu,power.draw", "--format=csv,noheader,nounits")
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
			if len(fields) >= 6 {
				memoryUsage, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					memoryUsage = 0
				}
				totalMemory, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					memoryUsage = 0
				}

				memoryUsageGiB := fmt.Sprintf("%.2f", memoryUsage/1024)
				totalMemoryGiB := fmt.Sprintf("%.2f", totalMemory/1024)

				gpu := GPUData{
					ID:          fields[0],
					Temperature: fields[1] + "°C",
					MemoryUsage: memoryUsageGiB + " / " + totalMemoryGiB + " GiB",
					GPUUtil:     fields[4] + "%",
					Power:       fields[5] + "W",
				}
				gpuData = append(gpuData, gpu)
			}
		}
	}
	return gpuData
}

func getAMDGPUData() ([]GPUData, error) {
	cmd := exec.Command("rocm-smi", "--showtemp", "--showuse", "--showpower", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var data map[string]map[string]string
	err = json.Unmarshal(output, &data)
	if err != nil {
		return nil, err
	}

	// Get memory usage data
	memCmd := exec.Command("rocm-smi", "--showmeminfo", "vram", "--json")
	memOutput, memErr := memCmd.Output()
	if memErr != nil {
		return nil, memErr
	}

	var memData map[string]map[string]string
	err = json.Unmarshal(memOutput, &memData)
	if err != nil {
		return nil, err
	}

	var gpuData []GPUData
	for card, info := range data {
		// Convert memory usage to GB
		usedMemBytes := memData[card]["VRAM Total Used Memory (B)"]
		totalMemBytes := memData[card]["VRAM Total Memory (B)"]
		usedMemGB, totalMemGB := convertBytesToGB(usedMemBytes, totalMemBytes)

		gpu := GPUData{
			ID:          card,
			Temperature: info["Temperature (Sensor edge) (C)"] + "°C",
			GPUUtil:     info["GPU use (%)"] + "%",
			Power:       info["Average Graphics Package Power (W)"] + "W",
			MemoryUsage: fmt.Sprintf("%s / %s GiB", usedMemGB, totalMemGB),
		}
		gpuData = append(gpuData, gpu)
	}

	return gpuData, nil
}

func convertBytesToGB(usedMemBytes, totalMemBytes string) (string, string) {
	usedMem, err := strconv.ParseFloat(usedMemBytes, 64)
	if err != nil {
		return "N/A", "N/A"
	}
	totalMem, err := strconv.ParseFloat(totalMemBytes, 64)
	if err != nil {
		return "N/A", "N/A"
	}

	usedMemGB := usedMem / (1024 * 1024 * 1024)
	totalMemGB := totalMem / (1024 * 1024 * 1024)

	return fmt.Sprintf("%.2f", usedMemGB), fmt.Sprintf("%.2f", totalMemGB)
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

type GPUResponse struct {
	Type     string    `json:"type"`
	GPUs     []GPUData `json:"GPUs"`
	Hostname string    `json:"hostname"`
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

	hostname, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GPUResponse{
		Type:     brand,
		GPUs:     gpuData,
		Hostname: hostname,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	http.HandleFunc("/gpustats", gpuHandler)
	http.ListenAndServe(":8998", nil)
}
