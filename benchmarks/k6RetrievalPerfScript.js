import http from 'k6/http';
import { sleep } from 'k6';

sleep(3);

function generateRandomString(length) {
  const characters = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  let randomString = '';

  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * characters.length);
    randomString += characters.charAt(randomIndex);
  }

  return randomString;
}

export default function () {
  const base_url = 'http://' + __ENV.HOSTNAME + '/api/2.0/mlflow/';


  // create experiments and runs
  let runIds = []
  let experimentIds = []
  // We need to load data into the database to run the retreival tests
  for(let i=0; i<3; i++) {
    let experiment_response = http.post(
      base_url + 'experiments/create',
      JSON.stringify({
          name: `experiment_${generateRandomString(5)}`,
      }),
      {
        headers: {
          'Content-Type': 'application/json'
        },
      }
    );

    let experimentId = experiment_response.json().experiment_id
    experimentIds.push(experimentId);

    for (let j=0; j<3; j++) {
      const run_response = http.post(
        base_url + 'runs/create',
        JSON.stringify({
          experiment_id: `${experimentId}`,
          start_time: Date.now(),
          run_name: `run_${generateRandomString(5)}_${experimentId}`,
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
      let runId = run_response.json().run.info.run_id;
      runIds.push(runId)
    }
  }

  // create sets of params
  let params = []
  // create sets of metrics
  let metrics = []


  for (let id = 1; id <= 100; id++) {
    // add params
    params.push({
      key: `param${id}`,
      value: `${id * Math.random()}`,
    })
    // add metrics
    for (let step=1; step < 5; step++){
      metrics.push({
        key: `metric${id}`,
        value: id * step * Math.random(),
        timestamp: Date.now(),
        step: step
      })
  }
  }

  // log metrics and params on runs
  for(let id = 0; id < runIds.length; id++){
    http.post(
      base_url + 'runs/log-batch',
      JSON.stringify({
        run_id: runIds[id],
        metrics: metrics,
        params: params
      }),
      {
        headers: {
          'Content-Type': 'application/json'
        },
      }
    );
  }

  // test searching for experiments
  http.post(
    base_url + 'runs/search',
    JSON.stringify({
      experiment_ids: experimentIds[0],
      max_results: 10,
      filter:" metrics.metric0 > 1 and params.param0 > 1",
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'SearchRuns',
      },
    }
  );

  // test searching for runs
  http.post(
    base_url + 'experiments/search',
    JSON.stringify({
      max_results: 10,
      filter:"name LIKE 'run_%'  AND tags.key = 'mlflow.user' ",
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
      tags: {
        name: 'SearchExperiments',
      },
    }
  );

  // test getting metric history
  http.get(
    base_url + `runs/metrics/get-history?run_id=${runIds[0]}&metric_key=metric1`,
    {
      tags: {
        name: 'MetricHistory',
      },
    }
  );

  //TODO: test getting metric histories
}