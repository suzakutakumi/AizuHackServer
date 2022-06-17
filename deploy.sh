
. ./.env

gcloud functions deploy Data --set-env-vars id=$id,secret=$secret --runtime go116 --trigger-http --allow-unauthenticated
