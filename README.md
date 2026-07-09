# 概要
ImageFlux Live Streamingライブ配信での、配信者・視聴者の認証認可Webhook/APIスタブとして機能します。

# 構築手順
1. さくらのクラウドでコンテナレジストリを作成します。
2. dockerフォルダでdocker build、docker pushをして、コンテナレジストリにイメージをアップロードします。
3. ImageFlux Live StreamingのAPIトークンを取得します。
4. Terraformフォルダ内の変数ファイルに上記で取得した値を反映し、Terraform init、Terraform applyを実行します。
5. 出力されたAppRun共用型URLに各種パスを添え（下部の記事を参照）、ImageFlux Live Streamingのチャンネル作成API呼び出し時に設定します。

## Dockerソース
Go言語で記述しています。
IMAGEFLUX_ACCESS_TOKEN環境変数が必須で、ここにImageFlux Live StreamingのAPIトークンを設定します。
/healthで死活確認ができます。

## Terraformソース
AppRun共用型とモニタリングスイート（ログおよびメトリクス）を作成します。
モニタリングスイートのログ保持期間は40日に設定しているため、期間追課金は発生しません。

# 参考
- [Qiita記事](https://qiita.com/creatio313/items/1949b93a8b1adfa77eb4)
