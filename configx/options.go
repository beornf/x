package configx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/logrusx"

	"github.com/knadh/koanf"

	"github.com/ory/x/watcherx"
)

type (
	OptionModifier func(p *Provider)
)

func WithContext(ctx context.Context) OptionModifier {
	return func(p *Provider) {
		p.ctx = ctx
	}
}

func WithImmutables(immutables []string) OptionModifier {
	return func(p *Provider) {
		p.immutables = immutables
	}
}

func OmitKeysFromTracing(keys []string) OptionModifier {
	return func(p *Provider) {
		p.excludeFieldsFromTracing = keys
	}
}

func AttachWatcher(watcher func(event watcherx.Event, err error)) OptionModifier {
	return func(p *Provider) {
		p.onChanges = append(p.onChanges, watcher)
	}
}

func WithLogrusWatcher(l *logrusx.Logger) OptionModifier {
	return AttachWatcher(LogrusWatcher(l))
}

func LogrusWatcher(l *logrusx.Logger) func(e watcherx.Event, err error) {
	return func(e watcherx.Event, err error) {
		l.WithField("file", e.Source()).
			WithField("event", e).
			WithField("event_type", fmt.Sprintf("%T", e)).
			Info("A change to a configuration file was detected.")

		if et := new(jsonschema.ValidationError); errors.As(err, &et) {
			l.WithField("event", fmt.Sprintf("%#v", et)).
				Errorf("The changed configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting the process.")
		} else if et := new(ImmutableError); errors.As(err, &et) {
			l.WithError(err).
				WithField("key", et.Key).
				WithField("old_value", fmt.Sprintf("%v", et.From)).
				WithField("new_value", fmt.Sprintf("%v", et.To)).
				Errorf("A configuration value marked as immutable has changed. Rolling back to the last working configuration revision. To reload the values please restart the process.")
		} else if err != nil {
			l.WithError(err).Errorf("An error occurred while watching config file %s", e.Source())
		} else {
			l.WithField("file", e.Source()).
				WithField("event", e).
				WithField("event_type", fmt.Sprintf("%T", e)).
				Info("Configuration change processed successfully.")
		}
	}
}

func WithStderrValidationReporter() OptionModifier {
	return func(p *Provider) {
		p.onValidationError = func(k *koanf.Koanf, err error) {
			p.printHumanReadableValidationErrors(k, os.Stderr, err)
		}
	}
}

func WithStandardValidationReporter(w io.Writer) OptionModifier {
	return func(p *Provider) {
		p.onValidationError = func(k *koanf.Koanf, err error) {
			p.printHumanReadableValidationErrors(k, w, err)
		}
	}
}