# Bootstrap a btcd peer

## clone this project and prepare

```shell
git clone https://github.com/BitTraceProject/btcd.git
cd btcd 
```

## build docker container

```shell
bash build.sh
```

## modify tmpl files with the help of comments, note that the docker-compose tmpl file don't need to modify

```shell
cd deploy
export tmpl_dir=${HOME}/.bittrace/tmpl
mkdir -p ${tmpl_dir}
cp tmpl/.env.tmpl ${tmpl_dir}/
cp tmpl/btcd.conf.tmpl ${tmpl_dir}/
cp tmpl/docker-compose.yaml.tmpl ${tmpl_dir}/
nano ${tmpl_dir}/.env.tmpl
nano ${tmpl_dir}/tmpl/btcd.conf.tmpl
```

## run bootstrap.sh

```shell
bash bootstrap.sh
```

## run clean.sh when finish

```shell
bash clean.sh <CONTAINER_NAME you want to clean>
```
