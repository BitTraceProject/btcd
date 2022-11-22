# Bootstrap a btcd peer

## clone this project and prepare

```shell
git clone https://github.com/BitTraceProject/btcd.git
cd btcd/deploy
mkdir tmpl/
```

## modify `.env.tmpl` with the help of comments

```shell
cp example/.env.example tmpl/.env.tmpl
nano tmpl/.env.tmpl
```

## modify `btcd.conf.tmpl` with the help of comments

```shell
cp example/btcd.config.example tmpl/btcd.config.tmpl
nano tmpl/btcd.config.tmpl
```

## run bootstrap.sh

```shell
bash bootstrap.sh
```
