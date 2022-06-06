# Discord bot for Minecraft server Management

`discord_mcm`は、Minecraftサーバーの起動や終了、コマンド送信が可能なDiscord Botです。  
Java版、Bedrock版関わらず利用できます。

## 機能

- サーバー起動/終了
- サーバーコマンド送信
- サーバーステータス確認
- 定期実行

## 使い方

初回起動時には、設定ファイル`dmcm_config.yml`が生成されます。  
ファイルを編集して設定し、再度起動してください。

```
$ ./discord_mcm
Generated dmcm_config.yml. Edit it!
$ vi dmcm_config.yml
$ ./discord_mcm
```

## 設定ファイル

```yaml
token: your token                              # Discord Botのトークン
channelId: your channel ID                     # やり取りするDiscordチャンネルID
launchCommand: java -jar minecraft_server.jar  # サーバー起動コマンド
prefix: m!                                     # コマンドプレフィックス
schedules:                                     # 定期実行の設定 ない場合は[]を指定
  - type: mc                                   # minecraftへ送信: mc、ホストへ送信: host
    command: stop                              # コマンド内容
    datetime: 0 12 * * *                       # 「分 時 日 月 曜」または「@every (時間)」
  - type: host
    command: java -jar minecraft_server.jar
    datetime: 5 12 * * *
```

## Botコマンド

### サーバー起動

```
m!start
```

### サーバー終了

```
m!stop
```

### サーバーコマンド送信

```
m!cmd [command]
```

### サーバーステータス確認

```
m!status
```
