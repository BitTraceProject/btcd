# Bootstrap a btcd peer

## switch to root user and clone this project, then prepare

```shell
git clone https://github.com/BitTraceProject/btcd.git
cd btcd 
```

## modify tmpl files with the help of comments, note that the docker-compose tmpl file don't need to modify

```shell
cd deploy
export tmpl_dir=/root/.bittrace/tmpl
mkdir -p ${tmpl_dir}
cp tmpl/* ${tmpl_dir}/
cp tmpl/.env.tmpl ${tmpl_dir}/
nano ${tmpl_dir}/.env.tmpl ## modify
nano ${tmpl_dir}/btcd.conf.tmpl ## modify
```

## bootstrap or restart container

```shell
bash bootstrap.sh
```

or

```shell
bash restart.sh ${CONTAINER_NAME}
```