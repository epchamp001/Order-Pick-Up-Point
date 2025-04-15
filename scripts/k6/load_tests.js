import http from 'k6/http';
import { check, fail } from 'k6';
import { Rate } from 'k6/metrics';

export const errorRate = new Rate('http_req_failed');

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 1000,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 100,
            maxVUs: 500,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<100'], // 95% запросов < 100мс
        http_req_failed: ['rate<0.0001'], // < 0.01% ошибок
    },
};

const BASE_URL = 'http://localhost:8080';
const HEADERS = { 'Content-Type': 'application/json' };

export default function () {
    // Получение токена модератора
    const loginRes = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({ role: 'moderator' }), {
        headers: HEADERS,
    });

    const tokenOk = check(loginRes, {
        'login status 200': (r) => r.status === 200,
        'token present': (r) => !!r.json('token'),
    });

    errorRate.add(!tokenOk);
    if (!tokenOk) {
        fail('Failed to get token');
    }

    const token = loginRes.json('token');
    const authHeaders = {
        ...HEADERS,
        Authorization: `Bearer ${token}`,
    };

    // Создание ПВЗ
    const pvzRes = http.post(`${BASE_URL}/pvz`, JSON.stringify({ city: 'Moscow' }), {
        headers: authHeaders,
    });

    const pvzOk = check(pvzRes, {
        'pvz status 201': (r) => r.status === 201,
        'pvz id present': (r) => !!r.json('id'),
    });

    errorRate.add(!pvzOk);
    if (!pvzOk) {
        fail('Failed to create PVZ');
    }

    // Получение списка ПВЗ
    const getPvzRes = http.get(`${BASE_URL}/pvz?page=1&limit=10`, {
        headers: authHeaders,
    });

    const getPvzOk = check(getPvzRes, {
        'get pvz status 200': (r) => r.status === 200,
        'get pvz returns array': (r) => Array.isArray(r.json()),
    });

    errorRate.add(!getPvzOk);
    if (!getPvzOk) {
        fail('Failed to get PVZ list');
    }
    //
    // // Получение списка ПВЗ (оптимизированный метод)
    // const getOptimizedRes = http.get(`${BASE_URL}/pvz/optimized?page=1&limit=10`, {
    //     headers: authHeaders,
    // });
    //
    // const getOptimizedOk = check(getOptimizedRes, {
    //     'optimized get pvz status 200': (r) => r.status === 200,
    //     'optimized get pvz returns array': (r) => Array.isArray(r.json()),
    // });
    //
    // errorRate.add(!getOptimizedOk);
    // if (!getOptimizedOk) {
    //     fail('Failed to get optimized PVZ list');
    // }
}