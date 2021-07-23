FROM nginx:stable-alpine

RUN rm /etc/nginx/nginx.conf
COPY ./nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
