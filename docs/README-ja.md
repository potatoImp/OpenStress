[English](./README-en.md) |
[中文](./README-zh.md) |
[日本語](./README-ja.md)

<div align="center">
  <h1>OpenStress</h1>
  <p>Goでの高性能ストレステストフレームワーク</p>
  
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-CODE">
    <img alt="コードライセンス" src="https://img.shields.io/badge/Code_License-MIT-f5de53?&color=f5de53"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-MODEL">
    <img alt="モデルライセンス" src="https://img.shields.io/badge/Model_License-Model_Agreement-f5de53?&color=f5de53"/>
  </a>
  <a href="https://golang.org/doc/install">
    <img alt="Go バージョン" src="https://img.shields.io/badge/Go-%3E%3D%201.16-blue"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/releases">
    <img alt="GitHub リリース" src="https://img.shields.io/github/v/release/potatoImp/OpenStress?color=brightgreen"/>
  </a>
</div>

## 目次

1. [紹介](#紹介)
2. [特徴](#特徴)
3. [クイックスタート](#クイックスタート)
4. [評価結果](#評価結果)
5. [インストール](#インストール)
6. [ライセンス](#ライセンス)
7. [連絡先](#連絡先)

## 紹介

OpenStressは、Goでの並行アプリケーションの開発を簡素化するために設計されたオープンソースのタスク管理およびログフレームワークです。タスクを管理し、アプリケーションイベントをログに記録する効率的な方法を提供し、生産性と保守性を向上させたい開発者にとって理想的な選択です。

## 特徴

- [x] **並行タスク管理**: OpenStressは、ワーカースレッドのプールを作成および管理し、タスクを並行して効率的に実行できるようにします。これにより、リソースの利用を最大化し、アプリケーションのパフォーマンスを向上させることができます。

- **柔軟なログシステム**: 組み込みのログフレームワークは、さまざまなログレベル（INFO、WARN、ERROR、DEBUG）をサポートし、開発者がアプリケーションの動作を追跡し、問題を簡単に診断できるようにします。ログはカスタマイズ可能で、さまざまな出力に向けることができます。

- **エラーハンドリング**: OpenStressには、開発者がカスタムエラータイプを定義し、エラーステートを効果的に管理できる強力なエラーハンドリングメカニズムが含まれています。これにより、OpenStressを使用して構築されたアプリケーションの信頼性が向上します。

- **Webベースのタスク管理インターフェース**: OpenStressは、タスク管理のための視覚的なWebインターフェースを提供し、ユーザー認証とタスクの開始と停止のためのAPIエンドポイントを備えています。また、統合と使用を容易にするためにSwaggerドキュメントをサポートしています。

- **簡素化された開発プロセス**: 開発者は、コルーチンプールの作成や並行性のプレッシャーの処理を心配することなく、タスクの実行に専念できます。OpenStressは、標準的なテストデータの収集、出力、および報告を担当し、開発ワークフローを合理化します。

- **簡単な統合**: OpenStressは、シンプルさを念頭に置いて設計されており、既存のGoプロジェクトに簡単に統合できます。モジュラーアーキテクチャにより、他のライブラリやフレームワークとシームレスに使用できます。

- **オープンソース**: オープンソースプロジェクトとして、OpenStressはコミュニティの貢献とコラボレーションを奨励しています。開発者は、フレームワークを自由に使用、変更、および配布でき、革新と改善の環境を育成します。

## クイックスタート

OpenStressを始めるには、次の手順に従ってください：

1. **リポジトリをクローン**:
   ```bash
   git clone https://github.com/potatoImp/OpenStress.git
   cd OpenStress
2. **依存関係をインストール**:

   マシンにGoがインストールされていることを確認してください。次のコマンドを使用して必要な依存関係をインストールします:
   ```bash
   go mod tidy
3. **アプリケーションを実行**:
  
   メインアプリケーションを実行するには、次のコマンドを使用します:
   ```bash
   go run main.go


## 4. 評価結果
#### 標準ベンチマーク

<div align="center">


|  | Benchmark (Metric) | OpenStress | locust | jmeter | - | - |
|---|-------------------|----------|--------|-------------|---------------|---------|
| requestTotal | - | one million | one million | one million | Dense | MoE |
| QPS | - | **34575** | 4687 | 29111 | - | - |
| requestTime | - | **1.5ms** | 153ms | 6ms | - | - |
| MinRequestTime | - | <1ms | 5ms | <1ms | - | - |
| MaxRequestTime | - | 1150ms | 1336ms | 3210ms | - | - |
| SuccessRate | - | 100% | 100% | 100% | - | - |
| CPU | - | 100% | 100% | 100% | - | - |
| Summary Report | - | yes | yes | yes | - | - |
| HtmlReport | - | yes | yes | yes | - | - |
| AllTestData | - | yes | no | yes | - | - |
| Artificial Intelligence Analysis| - | yes | no | no | - | - |








</div>

> [!NOTE]
> 8c 16G
> For more evaluation details, please check our paper. 


## インストール

### 前提条件

- Go 1.16 or higher
- Redis（オプション、分散シナリオ用）

### インストール手順

1. リポジトリをクローン:
```bash
git clone https://github.com/potatoImp/OpenStress.git
cd OpenStress
```

## ライセンス
OpenStressはMITライセンスの下でライセンスされています。詳細については、LICENSEファイルを参照してください。

## 連絡先
お問い合わせやサポートが必要な場合は、リポジトリの問題セクションを通じて、または直接メンテナに連絡してください。


## 謝辞

このプロジェクトを可能にした以下のライブラリの作者と貢献者に感謝の意を表します：

- **go-echarts**: A powerful charting library for Go.
- **go-redis**: A Redis client for Go.
- **gokrb5**: A Go library for Kerberos authentication.
- **ants**: A high-performance goroutine pool for Go.
- **zap**: A fast, structured, and leveled logging library for Go.
- **lumberjack**: A log rolling package for Go.
- **yaml**: A YAML support library for Go.
- **xxhash**: A fast non-cryptographic hash algorithm.
