//go:build windows

/*
Copyright 2026 The Radius Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconciler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const envtestProcessStopTimeout = 10 * time.Second

func stopTestEnvironment(env *envtest.Environment) error {
	// controller-runtime v0.24 uses unsupported Unix signals on Windows and keeps its process
	// handles private. Limit the workaround to matching direct children of this test process so
	// concurrent envtest suites and unrelated processes cannot be terminated.
	processNames := map[string]struct{}{
		windowsExecutableName(env.ControlPlane.GetAPIServer().Path): {},
		windowsExecutableName(env.ControlPlane.Etcd.Path):           {},
	}
	processIDs, err := childProcessIDs(uint32(os.Getpid()), processNames)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to find envtest processes: %w", err), env.Stop())
	}

	var cleanupErrors []error
	for _, processID := range processIDs {
		if err := terminateProcess(processID); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	if err := env.Stop(); err != nil {
		cleanupErrors = append(cleanupErrors, err)
	}
	return errors.Join(cleanupErrors...)
}

func childProcessIDs(parentProcessID uint32, processNames map[string]struct{}) (_ []uint32, retErr error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		retErr = errors.Join(retErr, windows.CloseHandle(snapshot))
	}()

	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}

	var processIDs []uint32
	for {
		processName := windowsExecutableName(windows.UTF16ToString(entry.ExeFile[:]))
		if entry.ParentProcessID == parentProcessID {
			if _, ok := processNames[processName]; ok {
				processIDs = append(processIDs, entry.ProcessID)
			}
		}

		err := windows.Process32Next(snapshot, &entry)
		if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
			return processIDs, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func windowsExecutableName(path string) string {
	return strings.TrimSuffix(strings.ToLower(filepath.Base(path)), ".exe")
}

func terminateProcess(processID uint32) (retErr error) {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, processID)
	if err != nil {
		return fmt.Errorf("failed to open envtest process %d: %w", processID, err)
	}
	defer func() {
		retErr = errors.Join(retErr, windows.CloseHandle(handle))
	}()

	if err := windows.TerminateProcess(handle, 1); err != nil {
		return fmt.Errorf("failed to terminate envtest process %d: %w", processID, err)
	}
	waitResult, err := windows.WaitForSingleObject(handle, uint32(envtestProcessStopTimeout/time.Millisecond))
	if err != nil {
		return fmt.Errorf("failed waiting for envtest process %d to stop: %w", processID, err)
	}
	if waitResult != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("timed out waiting for envtest process %d to stop", processID)
	}
	return nil
}
