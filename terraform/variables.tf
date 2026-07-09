variable "access_token" {
  type        = string
  description = "さくらのクラウドプロジェクトのアクセストークン"
  sensitive   = true
}
variable "access_token_secret" {
  type        = string
  description = "さくらのクラウドプロジェクトのアクセストークンシークレット"
  sensitive   = true
}
variable "imageflux_access_token" {
  type        = string
  description = "ImageFluxのAPIトークン"
  sensitive   = true
}
variable "zone" {
  type        = string
  description = "リソースを作成するゾーン"
  default     = "is1c"
}
variable "container_registry_image" {
  type        = string
  description = "コンテナレジストリのイメージ"
  default     = "cr.sakuracr.jp/image:v1"
}
variable "container_registry_password_wo" {
  type        = string
  description = "コンテナレジストリのパスワード"
  sensitive   = true
}
variable "container_registry_server" {
  type        = string
  description = "コンテナレジストリのサーバ"
  default     = "cr.sakuracr.jp"
}
variable "container_registry_username" {
  type        = string
  description = "コンテナレジストリのユーザー名"
  default     = "container_user"
}