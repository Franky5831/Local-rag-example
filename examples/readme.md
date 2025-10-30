# This is an example on how to run a 100% private RAG system locally.
The examples uses Docker, if you happen to run it on a M series Mac your performance will be badly hurt;
Docker does not access gpus on M series macs, that results on models being ran on the cpu making the system very slow.
As a workaround you can run models outside of Docker directly on your Mac.

The example uses the `gemma3:270m` to respond and teh `nomic-embed-text:latest` as the embedding model.
You can find Models [here](https://ollama.com/search); In this example I used really small and efficent models, the biggest the parameters size the slower the model will run but it will be more precise.

## Rerequisites
- [Docker](https://www.docker.com/)

## Run it
```bash
docker-compose down
docker-compose up --build
```
<sup>It's going to take a while, it needs to download about .5GB of models from Ollama.</sup>

1. [main.go](./src/main.go) is called as the app entrypoint.
2. [seed.go](./src/seed.go) is called from main: it takes all the data from the data directory.
3. [retrive.go](./src/retrive.go) is calle from main: it asks some questions regarding the data that has just been seeded.

### Expected output
```
Question: what do I have in the fridge?
Answer: you have tomatoes, milk and eggs

Question: when did gundam first came out?
Answer: The first Mobile Suit Gundam series aired in Japan on April 7, 1979.
```
The data used to provide those answers can be found in the [data](./src/data/) files.
