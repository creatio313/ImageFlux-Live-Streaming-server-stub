resource "sakura_apprun_shared" "imageflux_live_streaming_stub" {
  name = "ImageFlux Live StreamingのWebhook/APIスタブ"

  max_scale       = 3
  min_scale       = 0
  port            = 8080
  timeout_seconds = 60

  components = [{
    name       = "ImageFlux Live StreamingのWebhook/APIスタブコンテナ"
    max_cpu    = "0.5"
    max_memory = "1Gi"
    deploy_source = {
      container_registry = {
        image               = var.container_registry_image
        password_wo         = var.container_registry_password_wo
        password_wo_version = 1
        server              = var.container_registry_server
        username            = var.container_registry_username
      }
    }
    env = [{
      key   = "IMAGEFLUX_ACCESS_TOKEN"
      value = var.imageflux_access_token
    }]
    probe = {
      http_get = {
        path = "/health"
        port = 8080
      }
    }
  }]
  traffics = [{
    version_index = 0
    percent       = 100
  }]
}
resource "sakura_monitoring_suite_log_storage" "ils_stub_log_storage" {
  name                  = "ImageFlux Live StreamingのWebhook/API"
  description           = "ImageFlux Live StreamingのWebhook/APIスタブのログを保存するためのログストレージ"
  classification        = "shared"
  is_system             = false
  retention_period_days = 40
}
resource "sakura_monitoring_suite_log_routing" "ils_stub_log_routing" {
  resource_id    = sakura_apprun_shared.imageflux_live_streaming_stub.resource_id
  storage_id     = sakura_monitoring_suite_log_storage.ils_stub_log_storage.id
  publisher_code = "apprun"
  variant        = "applicationlog"
}
resource "sakura_monitoring_suite_metric_storage" "ils_stub_metric_storage" {
  name        = "ImageFlux Live StreamingのWebhook/APIスタブ"
  description = "ImageFlux Live StreamingのWebhook/APIスタブのメトリクスを保存するためのメトリクスストレージ"
  is_system   = false
}
resource "sakura_monitoring_suite_metric_routing" "ils_stub_metric_routing" {
  resource_id    = sakura_apprun_shared.imageflux_live_streaming_stub.resource_id
  storage_id     = sakura_monitoring_suite_metric_storage.ils_stub_metric_storage.id
  publisher_code = "apprun"
  variant        = "applicationmetrics"
}