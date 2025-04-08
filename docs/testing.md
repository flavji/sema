# for repository
npm install -g firebase-tools
firebase setup:emulators:firestore


brew install --cask google-cloud-sdk
gcloud components install beta
gcloud components install cloud-firestore-emulator


gcloud beta emulators firestore start --host-port=localhost:8080



# for authetenticiatn 

firebase init emulators
firebase emulators:start --only auth --host 192.168.1.88 --port 9099 --project test-project


