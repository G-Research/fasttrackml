import http from 'k6/http';
import { sleep } from 'k6';

sleep(1);


export default function () {
  const base_url = 'http://' + __ENV.HOSTNAME + '/api/2.0/mlflow/';

  const run_response = http.post(
    base_url + 'runs/create',
    JSON.stringify({
      experiment_id: '0',
      start_time: Date.now(),
      tags: [
        {
          key: "mlflow.user",
          value: "k6"
        }
      ]
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
    }
  );
  // const run_id = run_response.json().run.info.run_id;

  // let params = []
  // for (let id = 1; id <= 4; id++) {
  //   params.push({
  //     key: `param${id}`,
  //     value: `${id * Math.random()}`,
  //   })
  // }
  // http.post(
  //   base_url + 'runs/log-batch',
  //   JSON.stringify({
  //     run_id: run_id,
  //     params: params
  //   }),
  //   {
  //     headers: {
  //       'Content-Type': 'application/json'
  //     },
  //   }
  // );

  // let metrics = [];
  // for (let step = 1; step <= 10000; step++) {
  //   for (let id = 1; id <= 4; id++) {
  //     metrics.push({
  //       key: `metric${id}`,
  //       value: id * step * Math.random(),
  //       timestamp: Date.now(),
  //       step: step
  //     })
  //   }
  // }

  // http.post(
  //   base_url + 'runs/log-batch',
  //   JSON.stringify({
  //     run_id: run_id,
  //     metrics: metrics
  //   }),
  //   {
  //     headers: {
  //       'Content-Type': 'application/json'
  //     },
  //   }
  // );

  // http.post(
  //   base_url + 'runs/update',
  //   JSON.stringify({
  //     run_id: run_id,
  //     end_time: Date.now(),
  //     status: 'FINISHED'
  //   }),
  //   {
  //     headers: {
  //       'Content-Type': 'application/json'
  //     },
  //   }
  // );
}