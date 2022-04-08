# Discord bot for Minecraft server Management

`discord_mcm`は、Minecraftサーバーの起動や終了、コマンド送信が可能なDiscord Botです。  
Java版、Bedrock版関わらず利用できます。

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

## 主な機能、未実装機能

- [x] サーバー起動
- [x] サーバー終了
- [x] サーバーコマンド送信
- [ ] サーバーステータス確認
- [ ] 定期実行
