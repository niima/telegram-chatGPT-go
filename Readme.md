ChatGPT


```
docker build . -t chatgpt
docker run -d --name chatgpt -e OPENAI_API_KEY=YYY -e BOT_TOKEN=YYY --restart=unless-stopped chatgpt /opt/chatgptgo
```