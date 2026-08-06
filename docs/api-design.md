# API 設計書

## ドキュメント情報

| 項目 | 内容 |
|------|------|
| プロジェクト名 | Mattermost Free Desk Reservation Plugin |
| 推奨リポジトリ名 | `mattermost-plugin-freedesk` |
| プラグイン ID | `com.github.taku-devjp.freedesk` |
| バージョン | 0.1.0 |
| 作成日 | 2026-08-06 |
| ステータス | 確定 |
| 参照 | [要件定義書](./requirements.md)、[DB 設計書](./database-design.md) |

---

## 1. 概要

### 1.1 目的

フリーデスク予約プラグインが v0.1.0（MVP）で提供する REST API および Web App 連携の仕様を定義する。

**v0.1.0 の UI 入口**

- App Bar ボタン → **モーダル**（予約タブ / 管理タブ）
- スラッシュコマンド・RHS は **v0.1.0 対象外**（v0.2 以降）

**確認ダイアログ（FR-047）**

- 予約作成・取消（本人・代理取消）は **Web Client 側** で確認ダイアログを表示し、ユーザーが OK した場合のみ API を呼び出す
- API 自体は idempotent な DELETE / 単発 POST とし、ダイアログ制御はフロントエンドの責務

### 1.2 ベース URL

```
{site_url}/plugins/{plugin_id}/api/v1
```

例:

```
https://mattermost.example.com/plugins/com.github.taku-devjp.freedesk/api/v1
```

### 1.3 認証・認可

| 項目 | 仕様 |
|------|------|
| 認証 | Mattermost セッション Cookie または `Authorization: Bearer {token}` |
| 必須ヘッダー | `Mattermost-User-ID`（Web Client からの呼び出し時はサーバーが付与） |
| 未認証 | `401 Unauthorized` |
| 権限不足 | `403 Forbidden` |

starter template の `MattermostAuthorizationRequired` ミドルウェアを全 API に適用する。

**ロールと API 権限（NFR-020〜NFR-022）**

| ロール | 説明 | API 権限 |
|--------|------|----------|
| 一般ユーザー | Mattermost 認証済みユーザー | マトリクス参照、自分の予約 CRUD |
| プラグイン管理者 | プラグイン設定 `PluginAdminUserIDs` で指定 | 上記 + デスク CRUD、他人の予約代理取消 |
| Mattermost System Admin | システム管理者 | プラグイン設定変更（System Console）。**デスク CRUD・予約代理取消の権限は持たない** |

### 1.4 日付・タイムゾーン

| 項目 | 仕様 |
|------|------|
| タイムゾーン | **Asia/Tokyo 固定**（FR-050）。API レスポンスの `timezone` は常に `"Asia/Tokyo"` |
| 「今日」 | Asia/Tokyo 0:00 境界（FR-051） |
| 予約可能日 | 今日 ≦ `reserve_date` ≦ 今日 + `MaxAdvanceDays`（FR-001、FR-011） |
| 取消可能日 | 今日 ≦ `reserve_date`（FR-002、FR-052） |
| 昨日以前 | マトリクス表示のみ。予約・取消 API は `400 DATE_OUT_OF_RANGE` |

### 1.5 共通レスポンス形式

**成功時**

```json
{
  "data": { ... }
}
```

**エラー時**

```json
{
  "error": {
    "code": "DESK_ALREADY_RESERVED",
    "message": "指定されたデスクは既に予約されています。",
    "details": {}
  }
}
```

### 1.6 共通 HTTP ステータス

| コード | 用途 |
|--------|------|
| 200 | 取得成功 |
| 201 | 作成成功 |
| 204 | 削除成功（ボディなし） |
| 400 | リクエスト不正 |
| 401 | 未認証 |
| 403 | 権限なし |
| 404 | リソース不存在 |
| 409 | 競合（二重予約等） |
| 500 | サーバーエラー |

### 1.7 エラーコード一覧

| コード | HTTP | 説明 |
|--------|------|------|
| `INVALID_REQUEST` | 400 | パラメータ不正 |
| `DESK_NOT_FOUND` | 404 | デスクが存在しない |
| `RESERVATION_NOT_FOUND` | 404 | 予約が存在しない |
| `DESK_ALREADY_RESERVED` | 409 | デスク・日付が既に予約済み（FR-003） |
| `USER_ALREADY_RESERVED` | 409 | ユーザーが同日に既に予約済み（FR-004） |
| `DESK_INACTIVE` | 400 | 無効化されたデスク |
| `DATE_OUT_OF_RANGE` | 400 | 予約可能期間外、または昨日以前 |
| `MONTH_OUT_OF_RANGE` | 400 | マトリクス表示対象月が当月より前（前月ページング不可） |
| `DESK_HAS_RESERVATIONS` | 409 | デスク削除時に未来予約が残っている |
| `FORBIDDEN` | 403 | 操作権限なし |
| `INTERNAL_ERROR` | 500 | 内部エラー |

---

## 2. API 一覧（v0.1.0）

| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| GET | `/matrix` | マトリクス表示用データ取得（1 か月分） | User |
| GET | `/reservations/mine` | 自分の予約一覧 | User |
| GET | `/reservations/{id}` | 予約詳細 | User |
| POST | `/reservations` | 予約作成 | User |
| DELETE | `/reservations/{id}` | 予約取消（本人 / 代理取消） | User / Plugin Admin |
| GET | `/desks` | デスク一覧 | User |
| POST | `/admin/desks` | デスク作成 | Plugin Admin |
| PUT | `/admin/desks/{id}` | デスク更新 | Plugin Admin |
| DELETE | `/admin/desks/{id}` | デスク削除（論理） | Plugin Admin |
| GET | `/locations` | 拠点一覧 | User |
| GET | `/config` | フロント向け設定取得 | User |

---

## 3. API 詳細

### 3.1 GET `/matrix`

指定期間のデスク × 日付マトリクスデータを返す。モーダル内予約タブのメイン表示用（FR-010、FR-045）。

**Query Parameters**

| 名前 | 型 | 必須 | 説明 |
|------|-----|------|------|
| `year` | int | No | 表示年（省略時: Asia/Tokyo の当年） |
| `month` | int (1-12) | No | 表示月（省略時: Asia/Tokyo の当月） |
| `location_id` | string | No | 拠点 ID（省略時: デフォルト拠点） |

**日付範囲の算出**

1. `year` / `month` が **当月（Asia/Tokyo）より前** の場合 → `400 MONTH_OUT_OF_RANGE`
2. 当該月の `month_start`（1 日）〜 `display_end`（`min(当該月末, bookable_until)`）を `dates` として生成
3. **当月**を表示する場合、`month_start` 〜 `today - 1 日` も `dates` に含める（表示のみ、FR-052）。予約・取消は不可
4. 当該月に表示対象日が 0 件の場合も空の `dates` を返す（ページング UI のエラー防止）

**ページング（FR-045 — 翌月のみ）**

| 項目 | 仕様 |
|------|------|
| 前月 | **v0.1.0 では不可**。`can_go_prev` は常に `false` |
| 翌月 | `bookable_until` を含む月まで進める。それ以降は `can_go_next = false` |
| 初回表示 | 当月（`year` / `month` 省略時） |

**Response 200**

```json
{
  "data": {
    "year": 2026,
    "month": 8,
    "timezone": "Asia/Tokyo",
    "today": "2026-08-06",
    "bookable_until": "2026-10-05",
    "can_go_prev": false,
    "can_go_next": true,
    "desks": [
      {
        "id": "desk001",
        "name": "社員フリー1",
        "sort_order": 0,
        "is_active": true
      },
      {
        "id": "desk002",
        "name": "社員フリー2",
        "sort_order": 1,
        "is_active": true
      }
    ],
    "dates": [
      "2026-08-01",
      "2026-08-02",
      "2026-08-03",
      "2026-08-04",
      "2026-08-05",
      "2026-08-06"
    ],
    "reservations": [
      {
        "id": "res456",
        "desk_id": "desk001",
        "user_id": "user789",
        "user_name": "田中 太郎",
        "reserve_date": "2026-08-06",
        "is_mine": true
      }
    ]
  }
}
```

**フィールド補足**

| フィールド | 説明 |
|------------|------|
| `user_name` | Mattermost **フルネーム**。未設定時はユーザー名（FR-010） |
| `is_mine` | ログインユーザーの予約か（FR-012 ハイライト用） |
| `desks` | 有効デスクのみ（`is_active = true`）。最大 7 列（FR-044） |
| `dates` | 当該月の表示対象日。当月の場合は月初〜昨日も含む（表示のみ、FR-052） |
| `can_go_prev` | 前月へ進めるか。v0.1.0 では常に `false` |
| `can_go_next` | 翌月へ進めるか（`bookable_until` の月まで） |

---

### 3.2 GET `/reservations/mine`

ログインユーザーの有効な予約一覧（今日以降）。

**Query Parameters**

| 名前 | 型 | 必須 | 説明 |
|------|-----|------|------|
| `limit` | int | No | 件数上限（デフォルト: 50） |

**Response 200**

```json
{
  "data": {
    "reservations": [
      {
        "id": "res456",
        "desk_id": "desk001",
        "desk_name": "社員フリー1",
        "user_id": "user789",
        "reserve_date": "2026-08-06",
        "create_at": 1722902400000
      }
    ]
  }
}
```

---

### 3.3 GET `/reservations/{id}`

予約詳細を返す。v0.1.0 のモーダル UI では `GET /matrix` のデータで足りるため、主にデバッグ・将来拡張用。

**権限**

| 条件 | 結果 |
|------|------|
| 予約者本人 | `200` |
| Plugin Admin | `200` |
| 上記以外 | `403 FORBIDDEN` |

**Response 200**

```json
{
  "data": {
    "id": "res456",
    "desk_id": "desk001",
    "desk_name": "社員フリー1",
    "user_id": "user789",
    "user_name": "田中 太郎",
    "reserve_date": "2026-08-06",
    "create_at": 1722902400000,
    "update_at": 1722902400000
  }
}
```

---

### 3.4 POST `/reservations`

新規予約を作成する。Web Client は呼び出し前に **予約確認ダイアログ**（FR-042、FR-047）を表示する。

**Request Body**

```json
{
  "desk_id": "desk001",
  "reserve_date": "2026-08-07"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `desk_id` | string | Yes | デスク ID |
| `reserve_date` | string | Yes | 予約日（YYYY-MM-DD、Asia/Tokyo） |

**バリデーション**

1. `reserve_date` が今日以降、かつ `bookable_until` 以内
2. デスクが存在し `is_active = true`
3. `OneDeskPerDay = true` 時、同一ユーザーの同日予約がないこと

**Response 201**

```json
{
  "data": {
    "id": "res789",
    "desk_id": "desk001",
    "desk_name": "社員フリー1",
    "user_id": "user789",
    "reserve_date": "2026-08-07",
    "create_at": 1722988800000
  }
}
```

**Response 409（二重予約）**

```json
{
  "error": {
    "code": "DESK_ALREADY_RESERVED",
    "message": "指定されたデスクは既に予約されています。",
    "details": {
      "desk_id": "desk001",
      "reserve_date": "2026-08-07"
    }
  }
}
```

**副作用（FR-020、FR-021）**

- `EnableNotifications = true` かつ `NotificationChannelID` が設定されている場合、設定チャンネルに投稿
- 投稿内容: 予約者名・デスク名・日付
- 通知失敗時も予約は成功（非同期投稿、NFR-012 で INFO ログ）

---

### 3.5 DELETE `/reservations/{id}`

予約を取消（論理削除）する。Web Client は呼び出し前に **取消確認ダイアログ**（FR-043、FR-031、FR-047）を表示する。

**権限（NFR-021）**

| 操作 | 許可ロール |
|------|------------|
| 本人の予約取消 | 予約者本人 |
| 代理取消 | プラグイン管理者 |

**Response 204**

（ボディなし）

**副作用（FR-020、FR-021）**

- 通知 ON かつチャンネル設定ありの場合、取消投稿を送信

---

### 3.6 GET `/desks`

有効デスク一覧。一般ユーザー向け（管理タブの全件取得は Plugin Admin が `include_inactive` を使用）。

**Query Parameters**

| 名前 | 型 | 必須 | 説明 |
|------|-----|------|------|
| `location_id` | string | No | 拠点でフィルタ |
| `include_inactive` | bool | No | 無効デスクを含む（**Plugin Admin のみ**、デフォルト: false） |

**Response 200**

```json
{
  "data": {
    "desks": [
      {
        "id": "desk001",
        "location_id": "loc001",
        "name": "社員フリー1",
        "sort_order": 0,
        "is_active": true
      }
    ]
  }
}
```

---

### 3.7 POST `/admin/desks`

**権限**: Plugin Admin

**Request Body**

```json
{
  "location_id": "loc001",
  "name": "フリーデスク8",
  "sort_order": 7,
  "is_active": true
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `location_id` | string | Yes | 拠点 ID |
| `name` | string | Yes | デスク名称 |
| `sort_order` | int | No | 表示順（デフォルト: 0） |
| `is_active` | bool | No | 有効フラグ（デフォルト: true） |

**Response 201**

```json
{
  "data": {
    "id": "desk008",
    "location_id": "loc001",
    "name": "フリーデスク8",
    "sort_order": 7,
    "is_active": true,
    "create_at": 1722902400000
  }
}
```

---

### 3.8 PUT `/admin/desks/{id}`

**権限**: Plugin Admin

管理タブから名称・表示順・有効/無効を更新（FR-030）。

**Request Body**（部分更新可）

```json
{
  "name": "社員フリー1（窓際）",
  "sort_order": 0,
  "is_active": false
}
```

**Response 200**

更新後のデスクオブジェクト。

**`is_active = false` に変更した場合**

- 当該デスク列はマトリクスから即時非表示になる
- 既存の未来予約は DB に残るがマトリクスには表示されない（[DB 設計書 §3.2](./database-design.md) 参照）

---

### 3.9 DELETE `/admin/desks/{id}`

**権限**: Plugin Admin

デスクを論理削除する。`reserve_date >= today` の有効予約が 1 件でも残る場合は `409 DESK_HAS_RESERVATIONS` を返す。

**Response 204**

---

### 3.10 GET `/locations`

**Response 200**

```json
{
  "data": {
    "locations": [
      {
        "id": "loc001",
        "name": "Default",
        "sort_order": 0
      }
    ]
  }
}
```

v0.1.0 は 1 拠点のみ。将来の多拠点対応（v0.3.0）まで UI では拠点選択を提供しない。

---

### 3.11 GET `/config`

モーダル UI 向けの公開設定および現在ユーザーの権限情報。

**Response 200**

```json
{
  "data": {
    "timezone": "Asia/Tokyo",
    "today": "2026-08-06",
    "max_advance_days": 60,
    "bookable_until": "2026-10-05",
    "one_desk_per_day": true,
    "notification_enabled": true,
    "is_plugin_admin": false
  }
}
```

| フィールド | 説明 |
|------------|------|
| `timezone` | 常に `"Asia/Tokyo"`（読み取り専用、FR-050） |
| `max_advance_days` | 予約可能日数（System Admin が設定変更可、FR-011） |
| `bookable_until` | 予約可能最終日（`today + max_advance_days`） |
| `one_desk_per_day` | 1 日 1 デスク制限（FR-004、デフォルト `true`） |
| `notification_enabled` | 通知 ON/OFF（FR-020） |
| `is_plugin_admin` | 管理タブ表示可否（FR-046） |

---

## 4. Web App 連携（v0.1.0 — モーダル UI）

### 4.1 App Bar ボタン → モーダル（FR-040）

| 項目 | 仕様 |
|------|------|
| 登録 | Mattermost v9.0+ の App Bar Plugin API（`registerAppBarComponent` 等。実装時に starter template / 公式ドキュメントで API 名を確認） |
| 表示 | 画面右上 App Bar に常時表示。通知先チャンネル設定有無に **非依存** |
| トリガー | ボタンクリックでモーダルを開く |
| 閉じる | × ボタン / Esc / 背景クリック |

```typescript
registry.registerAppBarComponent(
  icon,
  () => store.dispatch(openFreeDeskModal()),
  'フリーデスク予約',
);
```

### 4.2 モーダル構成

| タブ | 表示条件 | 内容 |
|------|----------|------|
| 予約タブ | 全ユーザー | 日付 × デスク マトリクス（FR-010） |
| 管理タブ | `is_plugin_admin = true` のみ（FR-046） | デスク CRUD（FR-030） |

**初回表示フロー**

1. `GET /config` — 設定・権限取得
2. `GET /matrix?year=&month=` — 当月マトリクス取得
3. **翌月**ボタンで `year` / `month` を進めて再取得（FR-045）。前月ボタンは表示しない

### 4.3 マトリクス操作と API 呼び出し

いずれも **確認ダイアログ（FR-047）確定後** に API を呼ぶ。

| セル状態 | 操作 | API |
|----------|------|-----|
| 空き（今日以降） | 予約確認ダイアログ → OK | `POST /reservations` |
| 自分の予約（今日以降） | 取消確認ダイアログ → OK | `DELETE /reservations/{id}` |
| 他人の予約（Plugin Admin） | 代理取消確認ダイアログ → OK | `DELETE /reservations/{id}` |
| 昨日以前 | 操作不可 | — |

**確認ダイアログ表示内容（例 — FR-047）**

| 操作 | 表示内容 |
|------|----------|
| 予約 | デスク名、日付、「予約しますか？」 |
| 取消（本人） | デスク名、日付、「予約を取り消しますか？」 |
| 代理取消 | デスク名、日付、予約者名、「代理で予約を取り消しますか？」 |

### 4.4 UI レイアウト要件

| 要件 ID | 実装方針 |
|---------|----------|
| FR-044 | モーダル幅 ≧ 画面 80%。7 列は横スクロールなしを目標 |
| FR-045 | 日付行は 1 か月ページング（**翌月のみ**、前月不可）。列ヘッダー（デスク名）固定 |
| FR-012 | `is_mine = true` のセルをハイライト |

### 4.5 v0.1.0 で採用しない UI

| 方式 | 理由 |
|------|------|
| 前月ページング | v0.1.0 では不要（当月以降のみ表示） |
| RHS（右パネル） | モーダルで代替（要件定義 2.2） |
| チャンネルヘッダーボタン | App Bar が唯一の入口 |
| スラッシュコマンド | v0.2.0 以降（FR-041） |

---

## 5. スラッシュコマンド（v0.2.0 以降・参考）

v0.1.0 では **実装しない**。要件定義書 6.6 の仕様を v0.2.0 で実装予定。

| コマンド | 説明 |
|----------|------|
| `/freedesk` | 予約 UI（モーダル）を開く |
| `/freedesk book <desk> <YYYY-MM-DD>` | 指定デスク・日付を予約 |
| `/freedesk cancel <YYYY-MM-DD>` | 指定日の自分の予約を取消 |
| `/freedesk list [YYYY-MM-DD]` | 指定日の全デスク状況を表示 |
| `/freedesk mine` | 自分の予約一覧を表示 |

---

## 6. プラグイン設定（System Console — FR-032）

`plugin.json` の `settings_schema` で定義。**設定変更は Mattermost System Admin のみ**。

| 設定キー | 型 | デフォルト | 説明 |
|----------|-----|------------|------|
| `NotificationChannelID` | text | `""` | 通知投稿先チャンネル ID（FR-032） |
| `EnableNotifications` | bool | `true` | 予約・取消時のチャンネル通知（FR-020） |
| `MaxAdvanceDays` | number | `60` | 予約可能日数（約 2 か月、FR-011） |
| `OneDeskPerDay` | bool | `true` | 1 日 1 デスク制限（FR-004）。変更時は [DB 設計書 §3.3](./database-design.md) に従いユニークインデックスを DROP / CREATE |
| `PluginAdminUserIDs` | text | `""` | プラグイン管理者ユーザー ID（カンマ区切り） |

**v0.1.0 に含めない設定**

| 設定キー | 理由 |
|----------|------|
| `Timezone` | Asia/Tokyo 固定（FR-050）。変更不可 |

---

## 7. 内部 API / サービス層

Go サーバー内部のレイヤ構成（starter template 拡張）。

```
server/
├── api.go              # HTTP ハンドラ
├── plugin.go           # ライフサイクル（OnActivate: マイグレーション + 7 席シード）
├── store/
│   └── sqlstore/       # DB アクセス
│       ├── migration.go
│       ├── desk.go
│       └── reservation.go
└── service/
    ├── reservation.go  # ビジネスロジック（日付検証・競合チェック）
    └── notification.go # チャンネル投稿（FR-020、FR-021）
```

### 7.1 ReservationService.Create

```
1. Asia/Tokyo で today / bookable_until を算出
2. 入力バリデーション（日付形式、期間内、デスク有効、昨日以前拒否）
3. OneDeskPerDay = true 時、同一 user_id + reserve_date の既存予約を確認
4. トランザクション開始
5. INSERT freedesk_reservations（NFR-002: UNIQUE 制約で先着 1 件のみ成功）
6. UNIQUE 違反 → 409 DESK_ALREADY_RESERVED / USER_ALREADY_RESERVED
7. コミット
8. 非同期でチャンネル通知（失敗しても予約は成功、NFR-012 で INFO ログ）
```

排他制御は DB ユニーク制約を主とし、必要に応じてトランザクション内の事前 SELECT で早期 reject する（要件 §12 の Compare-And-Set は UNIQUE 違反ハンドリング + トランザクションで充足）。

### 7.2 ReservationService.Delete

```
1. 予約取得。不存在 → 404
2. reserve_date >= today 検証。昨日以前 → 400 DATE_OUT_OF_RANGE
3. 権限検証: 本人 or Plugin Admin。それ以外 → 403
4. 論理削除
5. 非同期で取消通知（NFR-012 で INFO ログ）
```

### 7.3 管理操作ログ（NFR-012）

デスク CRUD（`POST/PUT/DELETE /admin/desks`）および代理取消も、予約 CRUD と同様に INFO レベルでログ出力する。

---

## 8. ルーティング定義（Go 案）

```go
func (p *Plugin) initRouter() *mux.Router {
    router := mux.NewRouter()
    router.Use(p.MattermostAuthorizationRequired)

    api := router.PathPrefix("/api/v1").Subrouter()

    // Authenticated user
    api.HandleFunc("/matrix", p.handleGetMatrix).Methods(http.MethodGet)
    api.HandleFunc("/reservations/mine", p.handleGetMyReservations).Methods(http.MethodGet)
    api.HandleFunc("/reservations/{id}", p.handleGetReservation).Methods(http.MethodGet)
    api.HandleFunc("/reservations", p.handleCreateReservation).Methods(http.MethodPost)
    api.HandleFunc("/reservations/{id}", p.handleDeleteReservation).Methods(http.MethodDelete)
    api.HandleFunc("/desks", p.handleGetDesks).Methods(http.MethodGet)
    api.HandleFunc("/locations", p.handleGetLocations).Methods(http.MethodGet)
    api.HandleFunc("/config", p.handleGetConfig).Methods(http.MethodGet)

    // Plugin Admin only
    admin := api.PathPrefix("/admin").Subrouter()
    admin.Use(p.PluginAdminAuthorizationRequired)
    admin.HandleFunc("/desks", p.handleCreateDesk).Methods(http.MethodPost)
    admin.HandleFunc("/desks/{id}", p.handleUpdateDesk).Methods(http.MethodPut)
    admin.HandleFunc("/desks/{id}", p.handleDeleteDesk).Methods(http.MethodDelete)

    return router
}
```

---

## 9. セキュリティ考慮

| 項目 | 対策 |
|------|------|
| CSRF | Mattermost Web Client 経由の Cookie 認証に依存 |
| 入力検証 | 日付・ID 形式のバリデーション、SQL インジェクション対策（プリペアドステートメント） |
| レート制限 | v0.2 以降で検討 |
| 情報漏洩 | 予約者 `user_id` は返却するが、email 等の PII は含めない |
| 代理取消 | Plugin Admin のみ。System Admin には取消権限を付与しない（NFR-021） |

---

## 10. 改訂履歴

| バージョン | 日付 | 変更内容 | 担当 |
|------------|------|----------|------|
| 0.1.0 | 2026-08-06 | 初版作成 | — |
