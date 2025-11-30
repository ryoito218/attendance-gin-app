# 勤怠管理アプリ 要件定義（簡易版）

## 1．目的
Go（Gin）と MySQL を用いて，シンプルな勤怠管理アプリを作成する．  
本番に近い開発フロー（Issue → ブランチ → PR → テスト → デプロイ）を体験することを目的とする．

## 2．前提
- ユーザーは 1 名（user_id = 1）固定．
- フロントは HTML／CSS／JavaScript で実装する．
- DB は MySQL（Docker コンテナ）を使用する．

## 3．機能要件（MVP）
### 3.1 出勤打刻
- 当日の勤怠レコードに現在時刻を clock_in として記録する．
- 同じ日に複数回出勤打刻した場合はエラーとする．

### 3.2 退勤打刻
- 当日の勤怠レコードに現在時刻を clock_out として記録する．
- 出勤前の退勤や二重退勤はエラーとする．

### 3.3 勤怠一覧
- 過去 N 日分の勤怠情報（work_date，clock_in，clock_out，勤務時間）を取得する．

## 4．API（概要）
### POST /api/clock-in
- 出勤打刻．  
- レスポンス例：  
  - `{ "status": "ok", "clock_in": "2025-01-01T09:00:00+09:00" }`

### POST /api/clock-out
- 退勤打刻．  
- レスポンス例：  
  - `{ "status": "ok", "clock_out": "2025-01-01T18:00:00+09:00" }`

### GET /api/attendance
- 勤怠一覧取得．  
- レスポンス例（簡易）：  
  - `[ { "work_date": "2025-01-01", "work_duration_minutes": 540 } ]`

## 5．DB 設計
### attendance テーブル
| カラム名 | 型 | 説明 |
|---------|---------|---------|
| id | INT AUTO_INCREMENT PK | レコード ID |
| user_id | INT | ユーザー ID |
| work_date | DATE | 日付 |
| clock_in | DATETIME NULL | 出勤 |
| clock_out | DATETIME NULL | 退勤 |

制約：  
- `UNIQUE KEY (user_id, work_date)`

## 6．非機能要件
- Docker で MySQL を起動する．
- GitHub Actions で `go test ./...` を実行する．
- デプロイはオンプレミスサーバで `git pull` と `docker compose up -d` により行う．

## 7．開発フロー
1．GitHub Issue を作成する．  
2．Issue からブランチを作成する．  
3．ローカルで実装し，`go test` を実行する．  
4．コミットメッセージに Issue 番号を含めて push する．  
5．PR を作成し，main にマージする．  
