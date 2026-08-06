# 設計ドキュメント

Mattermost フリーデスク予約プラグイン（`mattermost-plugin-freedesk`）の設計資料。  
プラグイン ID: `com.freedesk.mattermost` / 対象バージョン: **v0.1.0（MVP）** / ステータス: **確定**

## 読む順序

1. [要件定義書](./requirements.md) — 背景、機能要件、非機能要件
2. [DB 設計書](./database-design.md) — テーブル、DDL、クエリ
3. [API 設計書](./api-design.md) — REST API、Web App（モーダル UI）連携

## ドキュメント一覧

| ドキュメント | 説明 |
|--------------|------|
| [要件定義書](./requirements.md) | 背景、機能要件、非機能要件、リリース計画 |
| [DB 設計書](./database-design.md) | テーブル定義、ER 図、DDL、クエリ |
| [API 設計書](./api-design.md) | REST API、Web App 連携（スラッシュコマンドは v0.2 参考） |

## v0.1.0（MVP）スコープ

| 含む | 含まない（将来） |
|------|------------------|
| App Bar → モーダル（予約タブ / 管理タブ） | スラッシュコマンド（v0.2） |
| 7 席自動登録、日付 × デスク マトリクス | RHS（右パネル） |
| 予約・取消（確認ダイアログ必須）、チャンネル通知 | 半日予約・多拠点（v0.3） |
| Asia/Tokyo 固定、1 か月ページング（翌月のみ） | モバイル UI |

## 開発ベース

- [mattermost/mattermost-plugin-starter-template](https://github.com/mattermost/mattermost-plugin-starter-template)
- Mattermost Server **v9.0 以上**
