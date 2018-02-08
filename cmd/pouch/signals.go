/*
Copyright 2018 Tuenti Technologies S.L. All rights reserved.

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

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tuenti/pouch"
)

type SignalHandler struct {
	State *pouch.PouchState

	signals chan os.Signal
	stop    chan struct{}
}

func NewSignalHandler(p *pouch.PouchState) *SignalHandler {
	return &SignalHandler{
		State: p,
	}
}

func (h *SignalHandler) ShowSecretsInfo() {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%d secrets managed\n", len(h.State.Secrets)))

	for _, secret := range h.State.Secrets {
		ttu, known := secret.TimeToUpdate()
		ttuStr := ttu.Format(time.RFC1123)
		if !known {
			ttuStr = "unknown"
		}
		buffer.WriteString(fmt.Sprintf(" %s - TTU: %s, used in %d files\n",
			secret.Name, ttuStr, len(secret.FilesUsing)))
		for _, f := range secret.FilesUsing {
			buffer.WriteString(fmt.Sprintf(" - %s\n", f.Path))
		}
	}

	log.Println(buffer.String())
}

func (h *SignalHandler) Start() {
	h.signals = make(chan os.Signal, 1)
	h.stop = make(chan struct{}, 1)

	signal.Notify(h.signals, syscall.SIGUSR2)

	go func() {
		for {
			select {
			case s := <-h.signals:
				switch s {
				case syscall.SIGUSR2:
					h.ShowSecretsInfo()
				}
			case <-h.stop:
				signal.Stop(h.signals)
				close(h.signals)
				close(h.stop)
				return
			}
		}
	}()
}

func (h *SignalHandler) Stop() {
	h.stop <- struct{}{}
}
