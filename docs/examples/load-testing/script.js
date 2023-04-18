import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus : 10,
    duration: '30s',
};

export default function () {
    let response =  http.get('http://nginx-universal-chart');
    check(response, { "status is 200": (r) => r.status == 200 });
}
