import http from 'k6/http';
import { sleep } from 'k6';

sleep(3);



export default function () {
  // Set base url from environment variable, the variable ' -e HOSTNAME = xxxx ' must be added to command line argument
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
      tags: {
        name: 'CreateRun',
      },
    }
  );

  // retrieve run id
  const runId = run_response.json().run.info.run_id;

  // create sets of params
  let params10 = []
  let params100 = []
  let params500 = []


  // create sets of metrics
  let metrics10 = []
  let metrics100 = []
  let metrics500 = []


  for (let id = 1; id <= 500; id++) {
    if (id <= 10) {
      // add params
      params10.push({
        key: `param10-${id}`,
        value: `${id * Math.random()}`,
      })

      // add metrics
      for (let step = 1; step < 5; step++) {
        metrics10.push({
          key: `metric10-${id}`,
          value: id * step * Math.random(),
          timestamp: Date.now(),
          step: step
        })
      }
    }
    if (id <= 100) {
      params100.push({
        key: `param100-${id}`,
        value: `${id * Math.random()}`,
      })

      // add metrics
      for (let step = 1; step < 5; step++) {
        metrics100.push({
          key: `metric100-${id}`,
          value: id * step * Math.random(),
          timestamp: Date.now(),
          step: step
        })

      }
    }

    params500.push({
      key: `param500-${id}`,
      value: `${id * Math.random()}`,
    })

    // add metrics
    for (let step = 1; step < 3; step++) {
      metrics500.push({
        key: `metric500-${id}`,
        value: id * step * Math.random(),
        timestamp: Date.now(),
        step: step
      })
    }

  }

  // test logging single metric value
  http.post(
    base_url + 'runs/log-metric',
    JSON.stringify({
      run_id: runId,
      key: "metric1",
      value: Math.random(),
      timestamp: Date.now(),
      step: 0
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogMetricSingle',
      },
    }

  );

  //test logging metric only batch 10
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics10,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogMetricBatch10',
      },
    }
  );

  //test logging metric only batch 100
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics100,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogMetricBatch100',
      },
    }
  );

  // test logging metric only batch of 500
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics500,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogMetricBatch500',
      },
    }
  );

  //test logging params only batch 10
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      params: params10,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogParamBatch10',
      },
    }
  );

  //test logging params only batch 100
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      params: params100,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogParamBatch100',
      },
    }
  );

  // test update run status to complete
  http.post(
    base_url + 'runs/update',
    JSON.stringify({
      run_id: runId,
      end_time: Date.now(),
      status: 'FINISHED'
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'UpdateRun',
      },
    }
  );

}
