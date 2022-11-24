# Bootstrap a btcd peer

## clone this project and prepare

```shell
git clone https://github.com/BitTraceProject/btcd.git
cd btcd/deploy
export tmpl_dir=${HOME}/.bittrace/tmpl
mkdir ${tmpl_dir}
```

## modify tmpl files with the help of comments, note that the docker-compose tmpl file don't need to modify

```shell
cp tmpl/* ${tmpl_dir}/
nano .env.tmpl
nano tmpl/btcd.conf.tmpl
```

## run bootstrap.sh

```shell
bash bootstrap.sh
```

## run clean.sh when finish

```shell
bash clean.sh <CONTAINER_NAME you want to clean>
```
