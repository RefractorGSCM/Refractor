#!/bin/bash

# This is the quickstart script for Refractor.

# Use tput to determine the right sequences for bold/colored text
bold=$(tput bold)
reset=$(tput sgr 0)
red=$(tput setaf 1)
white=$(tput setaf 7)
bg_red=$(tput setab 1)
bg_black=$(tput setab 0)
yellow=$(tput setaf 3)
green=$(tput setaf 2)

clear -x
echo "${bold}WELCOME TO THE REFRACTOR QUICKSTART SCRIPT${reset}"
echo ""
echo "This script will collect some information from you and then automatically configure and deploy Refractor for you."
echo ""
echo "Please make sure you've read through the requirements page of the documentation to ensure you have satisfied all pre-requisites."
echo "${bold}https://refractor.dmas.dev/#/installation/requirements${reset}"
echo ""
echo "This script has two execution modes: test and deploy."
echo "${yellow}Please run through test mode first${reset} and only run deploy mode you're sure everything is working correctly."
echo ""
echo "Would you like to ${bold}test${reset} or ${bold}deploy${reset}?"
read -p "> (test/deploy): " run_mode

staging=1
function draw_heading() {
  if [ "$run_mode,," == "deploy,," ]; then
    staging=0
    echo "${bold}${white}${bg_red}  !!!  RUNNING IN DEPLOY MODE  !!!  ${reset}"
  else
    echo "${bold}${white}${bg_black}     RUNNING IN TEST MODE     ${reset}"
  fi
}

function reset_screen() {
  clear
  draw_heading
}

reset_screen

# Make sure docker-compose is installed
if ! [ -x "$(command -v docker-compose)" ]; then
  echo "Error: docker-compose is not installed">&2
  exit 1
fi

# Make sure Refractor-Svelte is present
if [ ! -d "./Refractor-Svelte" ]; then
  echo ""
  echo "${bold}${yellow}Refractor-Svelte was not found. Cloning it with git now...${reset}"
  git clone git@github.com:RefractorGSCM/Refractor-Svelte.git
fi

# Ensure the deploy folders are created
mkdir ./deploy/kratos ./deploy/postgres ./deploy/nginx ./deploy/svelte 2> /dev/null

initial_setup=false
# Check if the various config files exist. If they don't, then copy them from defaults.
if ! [ -f "./docker-compose.yml" ]; then
  echo "File: docker-compose.yml does not exist. Copying default..."
  cp ./default/docker/docker-compose.yml ./docker-compose.yml
  initial_setup=true
fi

if ! [ -f "./deploy/nginx/app.conf" ]; then
  echo "File: ./deploy/nginx/app.conf does not exist. Copying default..."
  cp ./default/nginx/app.conf ./deploy/nginx/app.conf
  initial_setup=true
fi

if ! [ -f "./deploy/postgres/init.sql" ]; then
  echo "File: ./deploy/postgres/init.sql does not exist. Copying default..."
  cp ./default/postgres/init.sql ./deploy/postgres/init.sql
  initial_setup=true
fi

if ! [ -f "./deploy/kratos/kratos.yml" ]; then
  echo "File: ./deploy/kratos/kratos.yml does not exist. Copying default..."
  cp ./default/kratos/kratos.yml ./deploy/kratos/kratos.yml
  initial_setup=true
fi

if ! [ -f "./Refractor-Svelte/.env.production" ]; then
  echo "File: ./Refractor-Svelte/.env.production does not exist. Copying default..."
  cp ./default/svelte/.env.production ./Refractor-Svelte/.env.production
  initial_setup=true
fi

# If this is the initial setup, create a .neversetup flag file. This is so that if the user exits out of the script before populating
# the required config fields, the script will know on next run that the files still contains default values that need to be replaced.
# This file is deleted once the user fills the required config values for the first time.
if "$initial_setup"; then
  touch ./.neversetup
fi

if [ -f "./.neversetup" ]; then
  initial_setup=true
fi

if ! cmp ./default/docker/docker-compose.yml ./docker-compose.yml || \
   ! cmp ./default/nginx/app.conf ./deploy/nginx/app.conf || \
   ! cmp ./default/postgres/init.sql ./deploy/postgres/init.sql || \
   ! cmp ./default/kratos/kratos.yml ./deploy/kratos/kratos.yml || \
   ! cmp ./default/svelte/.env.production ./Refractor-Svelte/.env.production || \
   "$initial_setup"; then
  echo ""

  if ! "$initial_setup"; then
    echo "${bold}${red}Your config files were previously modified.${reset}"
    echo "You can choose to overwrite your changes (o), quit the script (q) or continue with your current config files (c)."
    echo "Continuing with current config files could have strange effects unless you absolutely know what you're doing. Proceed with caution."
    echo ""
    read -p "${bold}What would you like to do?${reset} (o/q/c): " decision
  else
    # if this is the initial setup, we skip the overwrite prompt and just automatically trigger an overwrite of the files
    # since we know that the defaults are not valid configuration files for Refractor.
    decision="o"
  fi

  echo ""
  echo "${bold}DOMAIN SETUP${reset}"
  #domain="USER INPUT"
  read -p "Enter your domain (without http/https): " domain

  if [ "$decision,," == "o,," ]; then
    # Overwrite changes
    echo "Overwriting changes..."

    # Create backup folders
    mkdir ./deploy/backup 2> /dev/null
    mkdir ./deploy/backup/docker ./deploy/backup/nginx ./deploy/backup/kratos ./deploy/backup/postgres ./deploy/backup/svelte 2> /dev/null

    # nginx config
    rm ./deploy/backup/nginx/app.conf 2> /dev/null # delete old backup
    mv ./deploy/nginx/app.conf ./deploy/backup/nginx/app.conf # rename current app.conf to app.conf.bak
    cp ./default/nginx/app.conf ./deploy/nginx/app.conf # copy default file

    # docker compose file
    rm ./deploy/backup/docker/docker-compose.yml 2> /dev/null
    mv ./docker-compose.yml ./deploy/backup/docker/docker-compose.yml
    cp ./default/docker/docker-compose.yml ./docker-compose.yml

    # postgres file
    rm ./deploy/backup/postgres/init.sql 2> /dev/null
    mv ./deploy/postgres/init.sql ./deploy/backup/postgres/init.sql
    cp ./default/postgres/init.sql ./deploy/postgres/init.sql

    # kratos file
    rm ./deploy/backup/kratos/kratos.yml 2> /dev/null
    mv ./deploy/kratos/kratos.yml ./deploy/backup/kratos/kratos.yml
    cp ./default/kratos/kratos.yml ./deploy/kratos/kratos.yml

    # svelte env file
    rm ./deploy/backup/svelte/.env.production 2> /dev/null
    mv ./Refractor-Svelte/.env.production ./deploy/backup/svelte/.env.production
    cp ./default/svelte/.env.production ./Refractor-Svelte/.env.production
    echo ""

    echo ""
    echo "${bold}DATABASE SETUP${reset}"
    #db_uri="USER INPUT"
    echo "What do you want your database root user to be named?"
    read -p "> " db_user

    echo ""
    echo "What do you want your database root password to be?"
    read -p "> " db_password

    kratos_dsn="postgres:\/\/${db_user}:${db_password}\@postgresd:5432\/kratos\?sslmode=disable\&max_conns=20\&max_idle_conns=4"
    refractor_dsn="postgres:\/\/${db_user}:${db_password}\@postgresd:5432\/refractor\?sslmode=disable"

    echo ""
    echo ""
    echo "${bold}INITIAL USER ACCOUNT SETUP${reset}"
    #initial_email="USER INPUT"
    read -p "> Email: " initial_email
    #initial_username="USER INPUT"
    read -p "> Username: " initial_username

    echo ""
    echo ""
    echo "${bold}SMTP MAIL SERVER SETUP${reset}"
    echo "If anything requested of you seems unfamiliar, you should check out your SMTP provider's connection guides."
    echo ""
    #smtp_host="USER INPUT"
    read -p "> SMTP Host (without port): " smtp_host
    #smtp_port="USER INPUT"
    read -p "> SMTP Port: " smtp_port
    #smtp_user="USER INPUT"
    read -p "> SMTP User: " smtp_user
    #smtp_password="USER INPUT"
    read -p "> SMTP Password: " smtp_password
    #smtp_from="USER INPUT"
    read -p "> SMTP From Address (e.g noreply@${domain}): " smtp_from

    smtp_uri="smtp:\/\/${smtp_user}:${smtp_password}\@${smtp_host}:${smtp_port}"

    # Generate a random 32 byte string for the encryption key.
    encryption_key=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32 ; echo "")

    # Generate a random 32 byte string for the kratos cookie secret.
    cookie_secret=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32 ; echo "")

    echo ""
    echo ""
    echo "${bold}COMMUNITY SETUP${reset}"
    echo "What's your community's name?"
    read -p "> " community_name
    echo ""

    # Set up env variables for frontend svelte
    kratos_root="https:\/\/${domain}\/kp"
    auth_root="https:\/\/${domain}\/k"
    api_root="https:\/\/${domain}\/api\/v1"
    websocket_root="wss:\/\/${domain}\/ws"

    # Write out variables to file placeholders
    sed -ri "s/\{\{DOMAIN\}\}/${domain}/g"                             ./deploy/nginx/app.conf ./docker-compose.yml ./deploy/kratos/kratos.yml
    sed -ri "s/\{\{KRATOS_DSN\}\}/${kratos_dsn}/g"                     ./docker-compose.yml ./deploy/kratos/kratos.yml
    sed -ri "s/\{\{REFRACTOR_DSN\}\}/${refractor_dsn}/g"               ./docker-compose.yml
    sed -ri "s/\{\{DB_USER\}\}/${db_user}/g"                           ./docker-compose.yml
    sed -ri "s/\{\{DB_USER_PWD\}\}/${db_password}/g"                   ./docker-compose.yml
    sed -ri "s/\{\{INITIAL_USER_EMAIL\}\}/${initial_email}/g"          ./docker-compose.yml
    sed -ri "s/\{\{INITIAL_USER_USERNAME\}\}/${initial_username}/g"    ./docker-compose.yml
    sed -ri "s/\{\{SMTP_URI\}\}/${smtp_uri}/g"                         ./docker-compose.yml
    sed -ri "s/\{\{SMTP_FROM\}\}/${smtp_from}/g"                       ./docker-compose.yml
    sed -ri "s/\{\{ENCRYPTION_KEY\}\}/${encryption_key}/g"             ./docker-compose.yml
    sed -ri "s/\{\{COOKIE_SECRET\}\}/${cookie_secret}/g"               ./deploy/kratos/kratos.yml
    sed -ri "s/\{\{COMMUNITY_NAME\}\}/${community_name}/g"             ./Refractor-Svelte/.env.production
    sed -ri "s/\{\{KRATOS_ROOT\}\}/${kratos_root}/g"                   ./Refractor-Svelte/.env.production
    sed -ri "s/\{\{AUTH_ROOT\}\}/${auth_root}/g"                       ./Refractor-Svelte/.env.production
    sed -ri "s/\{\{API_ROOT\}\}/${api_root}/g"                         ./Refractor-Svelte/.env.production
    sed -ri "s/\{\{WEBSOCKET_ROOT\}\}/${websocket_root}/g"             ./Refractor-Svelte/.env.production

    # remove state flag file
    rm -f ./.neversetup

  elif [ "$decision,," == "q,," ]; then\
    echo "${red}Exiting...${reset}"
    exit
  else
    echo ""
    echo "${yellow}Continuing with existing config files${reset}"
  fi
fi

echo ""
echo "${bold}Please enter your email for domain renewal notices.${reset}"
read -p "> " email
echo ""
echo ""

rsa_key_size=4096
data_path="./data/certbot"

# Check for existing data
if [ -d "$data_path" ]; then
  echo "${bold}Existing certificate data found for ${domain}. If you continue, this data will be overwritten.${reset}"
  read -p "> Would you like to continue? (y/n): " decision
  if [ "$decision,," != "y,," ]; then
    echo "${red}Exiting...${reset}"
    exit
  fi
fi

# Get recommended TLS params
if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "Fetching recommended TLS parameters..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
  echo ""
fi

# Create dummy certificate so that nginx can start up.
# If we didn't do this, nginx would not be able to start properly with our config which would make it so we cant generate a certificate.
# Classic chicken or the egg first situation.
echo "Generating a dummy certificate for $domain..."
path="/etc/letsencrypt/live/$domain"
mkdir -p "$data_path/conf/live/$domain"
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml run --rm --entrypoint "\
  openssl req -x509 -nodes -newkey rsa:$rsa_key_size -days 1\
    -keyout '$path/privkey.pem' \
    -out '$path/fullchain.pem' \
    -subj '/CN=localhost'" certbot
echo

reset_screen
echo "${bold}The various docker containers will now be built.${reset}"
echo ""
echo "You will see large amount of console output you may or may not recognize. Please let the script run uninterrupted."
echo "It may take several minutes to complete."
echo ""
sleep 7 # sleep to give the user time to read the text above

# Create frontend first. We do this here to avoid circular dependencies in the docker-compose file since in the future we could have different
# frontends which could add new unknowns. This is the safer option.
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up --force-recreate --no-deps -d refractor-frontend

# Start nginx
echo "Starting nginx..."
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up --force-recreate -d nginx

# Now that nginx started, we delete the dummy certificate
echo "Deleting dummy certificate for $domain..."
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml run --rm --entrypoint "\
  rm -Rf /etc/letsencrypt/live/$domain && \
  rm -Rf /etc/letsencrypt/archive/$domain && \
  rm -Rf /etc/letsencrypt/renewal/$domain.conf" certbot
echo

# Determine staging arg to use
if [ "$staging" != "0" ]; then staging_arg="--staging"; fi

# Select appropriate email arg
case "$email" in
  "") email_arg="--register-unsafely-without-email" ;;
  *) email_arg="--email $email" ;;
esac
echo ""

# Request certificate from LetsEncrypt
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    $email_arg \
    -d $domain \
    --rsa-key-size $rsa_key_size \
    --agree-tos \
    --force-renewal" certbot
echo

# Restart nginx to use the newly generated certificate and connect to hosts
sleep 3
docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml restart nginx

echo "If you received a message from nginx mentioning an unknown or disconnected host ${bold}refractor${reset}, something prevented the backend from starting."
echo "You should run ${bold}docker logs refractor${reset} to see what the issue is."
echo "Once the issue was fixed and the backend is functioning normally, you must restart the proxy using ${bold}docker restart nginx${reset}"
echo ""
echo "If you didn't get any errors, then you're all set!"
echo ""
echo "Enjoy Refractor!"
echo ""