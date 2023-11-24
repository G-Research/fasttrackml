import http from 'k6/http';
import { sleep } from 'k6';

sleep(1);

export default function () {
  // Set base url from environment variable, the variable ' -e HOSTNAME = xxxx ' must be added to command line arguement
  const base_url = 'http://' + __ENV.HOSTNAME + '/api/2.0/mlflow/';

  // test creating runs
  const run_response = http.post(
    base_url ,
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

}