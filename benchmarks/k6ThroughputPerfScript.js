import http from 'k6/http';
import { sleep } from 'k6';

sleep(1);

export default function () {
  // Set base url from environment variable, the variable ' -e HOSTNAME = xxxx ' must be added to command line arguement
  const base_url = 'http://' + __ENV.HOSTNAME + '/api/2.0/mlflow/';

  // test creating runs
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
  let params5 = []
  let params10 = []
  let params100 = []


  // create sets of metrics
  let metrics5 = []
  let metrics10 = []
  let metrics100 = []


  for (let id = 1; id <= 100; id++) {
    if(id <= 5){
      // add params
      params5.push({
        key: `param${id}`,
        value: `${id * Math.random()}`,
      })

      // add metrics
      for (let step=1; step < 5; step++){
        metrics5.push({
          key: `metric${id}`,
          value: id * step * Math.random(),
          timestamp: Date.now(),
          step: step
        })
    }
  }
    if(id <= 10){
      // add params
      params10.push({
        key: `param${id}`,
        value: `${id * Math.random()}`,
      })

      // add metrics
      for (let step=1; step < 5; step++){
        metrics10.push({
          key: `metric${id}`,
          value: id * step * Math.random(),
          timestamp: Date.now(),
          step: step
        })
      }
    }

    params100.push({
      key: `param${id}`,
      value: `${id * Math.random()}`,
    })

    // add metrics
    for (let step=1; step < 5; step++){
      metrics100.push({
        key: `metric${id}`,
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
      key: "metric0",
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

  // test logging batch of 5
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics5,
      params: params5
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'LogMetricBatch5',
      },
    }
  );

  // test logging batch of 10
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics10,
      params: params10
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


  // test logging batch of 100
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: runId,
      metrics: metrics100,
      params: params100
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