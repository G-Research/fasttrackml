query_0 = 'run.hparams.benchmark == "glue" ' \
          'and run.hparams.dataset == "cola" ' \
          'and metric.context.subset != "train"'
query_1 = 'run.hparams.benchmark == "glue" ' \
          'and run.hparams.dataset == "cola"'
query_2 = 'run.hparams.benchmark == "glue"'
query_3 = 'run.hparams.dataset == "cola" ' \
          'and run.experiment.name != "baseline-warp_4-cola"'


queries = {
    # each query contains a tupple of 2 queries, the first being for aim, while the second being for mlflow and fasttrack
    0: (query_0, "metric.accuracy > 0"),
    1: (query_1, "metric.accuracy > 0"),
    2: (query_2, "metric.accuracy > 0"),
    3: (query_3, "metric.accuracy > 0"),
}

