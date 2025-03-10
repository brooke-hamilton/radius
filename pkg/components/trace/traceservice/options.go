/*
Copyright 2023 The Radius Authors.

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

package traceservice

// Options represents the trace options.
type Options struct {
	// Enabled configures whether tracing is enabled.
	Enabled bool `yaml:"enabled"`
	// ServiceName represents the name of service.
	ServiceName string `yaml:"serviceName,omitempty"`
	// Zipkin represents zipkin options.
	Zipkin *ZipkinOptions `yaml:"zipkin,omitempty"`
}

// ZipkinOptions represents zipkin trace provider options.
type ZipkinOptions struct {
	// URL represents the url of zipkin endpoint.
	URL string `yaml:"url"`
}
