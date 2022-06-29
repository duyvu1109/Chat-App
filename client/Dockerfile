FROM node:14.19

RUN mkdir -p /appclient

WORKDIR /appclient

COPY package*.json ./

RUN npm install

COPY ./client .

EXPOSE 3000

CMD [ "npm", "start" ]