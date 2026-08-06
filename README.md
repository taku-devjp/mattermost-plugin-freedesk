# Mattermost Free Desk Plugin

[![CI](https://github.com/taku-devjp/mattermost-plugin-freedesk/actions/workflows/ci.yml/badge.svg)](https://github.com/taku-devjp/mattermost-plugin-freedesk/actions/workflows/ci.yml)

Mattermost 上でフリーデスク（ホットデスク）の予約・確認・取消を行うプラグイン。Excel による手動運用を置き換える。

| 項目 | 値 |
|------|-----|
| プラグイン ID | `com.freedesk.mattermost` |
| 対象 Mattermost | v9.0 以上 |
| 現フェーズ | v0.1.0（MVP）設計完了 → 実装中 |

## 機能概要（v0.1.0）

- **App Bar** から **モーダル** を開き、日付 × デスク（7 席）のマトリクスで予約状況を確認
- 空きセルの予約・自分の予約の取消（確認ダイアログ必須）
- プラグイン管理者向け **管理タブ**（デスク CRUD、代理取消）
- 予約・取消時の **チャンネル通知**（設定可能）
- タイムゾーン **Asia/Tokyo** 固定、表示は **1 か月単位**（翌月ページングのみ）

詳細は [docs/requirements.md](./docs/requirements.md) を参照。

## 設計ドキュメント

[docs/README.md](./docs/README.md) に一覧と読む順序があります。

| ドキュメント | 内容 |
|--------------|------|
| [要件定義書](./docs/requirements.md) | 機能要件・非機能要件 |
| [DB 設計書](./docs/database-design.md) | スキーマ・DDL |
| [API 設計書](./docs/api-design.md) | REST API・Web App 連携 |

---

## プラグインの利用（管理者・エンドユーザー）

開発環境を持たない場合は、ビルド済みバンドルを Mattermost にアップロードする。

1. **System Console** → **プラグイン** → **プラグイン管理**
2. `dist/com.freedesk.mattermost.tar.gz` をアップロード（または GitHub Releases の成果物）
3. プラグインを **有効化**
4. 設定（通知先チャンネル ID、プラグイン管理者ユーザー ID 等）を入力

---

## 開発

### 前提条件

- **Go**（`go.mod` 準拠）
- **Node.js** — リポジトリルートで `nvm use`（[`.nvmrc`](./.nvmrc) 参照）
- **make**

### Mattermost 側の準備

`config.json` でプラグインアップロードを有効化し、サーバーを再起動する。

```json
"PluginSettings": {
    "EnableUploads": true
}
```

### 環境変数

`.env.example` をコピーして `.env` を作成するか、シェルで export する。

| 変数 | 説明 |
|------|------|
| `MM_SERVICESETTINGS_SITEURL` | Mattermost の URL（必須） |
| `MM_ADMIN_TOKEN` | 管理者 Personal Access Token（推奨） |
| `MM_ADMIN_USERNAME` / `MM_ADMIN_PASSWORD` | トークンの代わりにログイン認証も可 |

ローカル Mattermost で **local mode** が有効な場合、上記なしで `make deploy` も可能（Unix ソケット経由）。

### コマンド

| コマンド | 用途 |
|----------|------|
| `make` | ビルド（`dist/com.freedesk.mattermost.tar.gz` を生成） |
| `make test` | テスト実行（PR 前） |
| `make deploy` | ビルド → アップロード → **有効化**（開発サーバー向け） |
| `make watch` | Web App の変更を監視して再ビルド（別ターミナル） |
| `make deploy-from-watch` | watch 中の Web App 変更をサーバーへ反映 |
| `make reset` | プラグイン無効化 → 再有効化（挙動がおかしいとき） |
| `make clean` | ビルド成果物の削除 |

**初回デプロイ**

```bash
cp .env.example .env   # 必要に応じて編集
make deploy
```

**Web App を編集しながら開発（2 ターミナル）**

```bash
# ターミナル 1
make watch

# ターミナル 2 — webapp 再ビルド後に実行
make deploy-from-watch
```

Go（server）を変更した場合は `make deploy` を再実行する。

---

## リポジトリ構成

```
server/     Go プラグイン（API、DB、ビジネスロジック）
webapp/     React / TypeScript（モーダル UI）
docs/       要件定義・設計書
plugin.json プラグインマニフェスト
```

## ライセンス

[mattermost-plugin-starter-template](https://github.com/mattermost/mattermost-plugin-starter-template) をベースに開発。
