import http from 'k6/http';

export default function () {
  const namespace = 'default'
  const numberOfExperiments = 2
  const runsPerExperiment = 10
  const paramsPerRun = 100
  const metricsPerRun = 1000
  const stepsPerMetric = 10

  for (let i = 0; i < numberOfExperiments; i++) {
    const experimentId = createExperiment(namespace)
    for (let j = 0; j < runsPerExperiment; j++) {
      createRun(namespace, experimentId, paramsPerRun, metricsPerRun, stepsPerMetric)
    }
  }
}

function createExperiment(namespace) {  
  const base_url = `http://localhost:5000/ns/${namespace}/api/2.0/mlflow/`;
  
  const exp_response = http.post(
    base_url + 'experiments/create',
    JSON.stringify({
      "name": `experiment-${Date.now()}`,
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
    }
  );
  return exp_response.json().experiment_id;
}


function createRun(namespace, experimentId, numParams, numMetrics, numSteps) {  
  const base_url = `http://localhost:5000/ns/${namespace}/api/2.0/mlflow/`;
  
  const run_response = http.post(
    base_url + 'runs/create',
    JSON.stringify({
      experiment_id: experimentId,
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
  const run_id = run_response.json().run.info.run_id;

  let params = []
  for (let id = 1; id <= numParams; id++) {
    params.push({
      key: `param${id}`,
      value: `${id * Math.random()}`,
    })
  }
  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: run_id,
      params: params
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
    }
  );

  let metrics = [];
  for (let step = 1; step <= numSteps; step++) {
    for (let id = 1; id <= numMetrics; id++) {
      metrics.push({
        key: `metric${id}`,
        value: id * step * Math.random(),
        timestamp: Date.now(),
        step: step,
      })
    }
  }

  http.post(
    base_url + 'runs/log-batch',
    JSON.stringify({
      run_id: run_id,
      metrics: metrics
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
    }
  );

  http.post(
    base_url + 'runs/update',
    JSON.stringify({
      run_id: run_id,
      end_time: Date.now(),
      status: 'FINISHED'
    }),
    {
      headers: {
        'Content-Type': 'application/json'
      },
    }
  );
}
