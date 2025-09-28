package tracing

import (
	"github.com/uber/jaeger-client-go/config"
)

const serviceName = "cart"

func Init(address string) error {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: address,
		},
	}

	_, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		return err
	}

	return nil
}
