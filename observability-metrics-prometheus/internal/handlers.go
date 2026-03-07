// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/openchoreo/community-modules/observability-metrics-prometheus/internal/api/gen"
	"github.com/openchoreo/community-modules/observability-metrics-prometheus/internal/observer"
)

const (
	alertAnnotationRuleName      = "rule_name"
	alertAnnotationRuleNamespace = "rule_namespace"
	alertAnnotationAlertValue    = "alert_value"
)

type MetricsHandler struct {
	observerClient *observer.Client
	logger         *slog.Logger
}

func NewMetricsHandler(observerClient *observer.Client, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		observerClient: observerClient,
		logger:         logger,
	}
}

var _ gen.StrictServerInterface = (*MetricsHandler)(nil)

func (h *MetricsHandler) Health(context.Context, gen.HealthRequestObject) (gen.HealthResponseObject, error) {
	return gen.Health200JSONResponse{
		Status: "healthy",
	}, nil
}

func (h *MetricsHandler) HandleAlertmanagerWebhook(
	ctx context.Context,
	request gen.HandleAlertmanagerWebhookRequestObject,
) (gen.HandleAlertmanagerWebhookResponseObject, error) {
	if request.Body == nil {
		return gen.HandleAlertmanagerWebhook400JSONResponse{
			Title:   "badRequest",
			Message: "request body is required",
		}, nil
	}

	if len(request.Body.Alerts) != 1 {
		return gen.HandleAlertmanagerWebhook400JSONResponse{
			Title:   "badRequest",
			Message: "expected exactly one alert in the webhook payload",
		}, nil
	}

	alert := request.Body.Alerts[0]
	if strings.ToLower(strings.TrimSpace(alert.Status)) != "firing" {
		return gen.HandleAlertmanagerWebhook400JSONResponse{
			Title:   "badRequest",
			Message: "expected a firing alert",
		}, nil
	}

	ruleName, ruleNamespace, alertValue, err := extractAlertFields(alert)
	if err != nil {
		return gen.HandleAlertmanagerWebhook400JSONResponse{
			Title:   "badRequest",
			Message: fmt.Sprintf("invalid firing alert: %v", err),
		}, nil
	}

	if err := h.observerClient.ForwardAlert(ctx, ruleName, ruleNamespace, alertValue, alert.StartsAt); err != nil {
		h.logger.Error("Failed to forward alert webhook to observer API",
			slog.String("ruleName", ruleName),
			slog.String("ruleNamespace", ruleNamespace),
			slog.Any("error", err),
		)
		return gen.HandleAlertmanagerWebhook500JSONResponse{
			Title:   "internalServerError",
			Message: "failed to forward alert webhook",
		}, nil
	}

	message := "processed webhook successfully; forwarded 1 firing alert"
	return gen.HandleAlertmanagerWebhook200JSONResponse{
		Status:  gen.Success,
		Message: message,
	}, nil
}

func extractAlertFields(alert gen.AlertmanagerAlert) (string, string, float32, error) {
	ruleName := strings.TrimSpace(alert.Annotations.RuleName)
	if ruleName == "" {
		return "", "", 0, fmt.Errorf("missing %q annotation", alertAnnotationRuleName)
	}

	ruleNamespace := strings.TrimSpace(alert.Annotations.RuleNamespace)
	if ruleNamespace == "" {
		return "", "", 0, fmt.Errorf("missing %q annotation", alertAnnotationRuleNamespace)
	}

	alertValueText := strings.TrimSpace(alert.Annotations.AlertValue)
	if alertValueText == "" {
		return "", "", 0, fmt.Errorf("missing %q annotation", alertAnnotationAlertValue)
	}

	alertValue, err := strconv.ParseFloat(alertValueText, 32)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid %q annotation: %w", alertAnnotationAlertValue, err)
	}

	return ruleName, ruleNamespace, float32(alertValue), nil
}
