#!/bin/bash

openssl aes-256-cbc -K $encrypted_45b586f32ae3_key -iv $encrypted_45b586f32ae3_iv -in .dockercfg.enc -out ~/.dockercfg -d

if [ "$TRAVIS_BRANCH" == "master" ]; then
  echo "Deploying image to docker hub for master (latest)"
  docker push "${TRAVIS_REPO_SLUG}:latest"
elif [ ! -z "$TRAVIS_TAG" ] && [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  echo "Deploying image to docker hub for tag ${TRAVIS_TAG}"
  docker push "${TRAVIS_REPO_SLUG}:${TRAVIS_TAG}"
elif [ ! -z "$TRAVIS_BRANCH" ] && [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  echo "Deploying image to docker hub for branch ${TRAVIS_BRANCH}"
  docker push "${TRAVIS_REPO_SLUG}:${TRAVIS_BRANCH}"
else
  echo "Not deploying image"
fi
