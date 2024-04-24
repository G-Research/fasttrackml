import argparse
import glob
import logging
import os

import matplotlib.pyplot as plt
import pandas as pd

BENCHMARKS = [
    "SearchRuns",
    "SearchExperiments",
    "MetricHistory",
    "CreateRun",
    # "LogMetricSingle",
    "LogMetricBatch10",
    "LogMetricBatch100",
    # "LogMetricBatch500"
    "LogParamBatch10",
    "LogParamBatch100",
]

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)


def generateReport(dfs, filename):
    """
    Generate an image report for a given dataframe and storing it with the
    provided filename.
    """
    # Sample data
    colors1 = ["red", "green", "blue", "orange"]

    num_dfs = len(dfs)
    num_cols = int(num_dfs / 2)

    # Create a figure with subplots
    fig, axes = plt.subplots(nrows=2, ncols=num_cols, figsize=(30, 20))

    # generate bar charts from dataframes
    for i in range(num_dfs):
        dfs[i].plot(
            kind="bar",
            x="application",
            y="metric_value",
            ax=axes[int(i / num_cols)][i % num_cols],
            legend=False,
            color=colors1,
        )

    # Customize the layout, labels, and title
    plt.subplots_adjust(wspace=0.5, hspace=0.5)  # Adjust the space between subplots

    for i in range(num_dfs):
        axes[int(i / num_cols)][i % num_cols].set_xlabel("Application")
        axes[int(i / num_cols)][i % num_cols].set_ylabel("Milliseconds")
        axes[int(i / num_cols)][i % num_cols].set_title(dfs[i]["name"][0])

    # Save the figure to a single image file (e.g., PNG)
    plt.savefig('performanceReport.png')




def getDataframeFromFile(filename, application_name):
    """
    Generate a dataframe object from reading a file
    and add a column to the dataframe object to represent what application the file is from
    """

    # Read the CSV file into a DataFrame
    df = pd.read_csv(filename)

    # Display the DataFrame
    df = df[df["metric_name"] == "http_req_duration"]
    df = df[df["name"].isin(BENCHMARKS)]

    df = df[df["metric_name"] == "http_req_duration"]
    df = df[df["name"].isin(BENCHMARKS)]

    # List of columns to keep
    columns_to_keep = ["metric_value", "name"]

    # Drop all columns except the specified ones
    df.drop(df.columns.difference(columns_to_keep), axis=1, inplace=True)
    df["application"] = application_name
    return df


def extractApplicationName(name: str):
    """
    This function is used to extract the application name from a generated
    file. e.g 'benchmark_outputs/mlflow_sqlite_logging.csv' would be 'mlfow sqlite'
    """
    name = name.split("/")[-1]
    name = name.split(".")[0]
    application_name = " ".join(name.split("_")[:-1])
    return application_name


def generateDataframes():
    """
    Generate single dataframe by concatenating the results from the various report files
    Filter the dataframe for only rows with benchmarks we want to measure
    """

    dataframes = []  # store the list of dataframes from all generated report files

    files = glob.glob("benchmark_outputs/*.csv")
    for file in files:
        dataframes.append(getDataframeFromFile(file, extractApplicationName(file)))
    # Read the CSV file into a DataFrame
    df = pd.concat(dataframes, ignore_index=True)

    dfs = []
    for benchmark in BENCHMARKS:
        benchmark_df = df[df["name"] == benchmark]
        benchmark_df = benchmark_df.groupby("application")["metric_value"].mean().reset_index()
        benchmark_df["name"] = benchmark
        dfs.append(benchmark_df)

    return dfs


def checkAllFilesReady():
    """
    This is used to check whether all the required output files have been generated
    Since the K6 tests would be run inside containers before shutting down, we need to check if their
    execution is complete before starting the report generataiton.
    The files we are checking to ensure they exist are:
    - mlflowsqlitelogging.csv
    - mlflowpostgreslogging.csv
    - fasttracksqlitelogging.csv
    - fasttrackpostgreslogging.csv
    - mlflowsqliteretrieval.csv
    - mlflowpostgresretrieval.csv
    - fasttrackpostgresretrieval.csv
    - fasttracksqliteretrieval.csv
    """
    if os.path.exists("benchmark_outputs/mlflowsqlitelogging.csv") and \
       os.path.exists("benchmark_outputs/mlflowpostgreslogging.csv") and \
       os.path.exists("benchmark_outputs/fasttracksqlitelogging.csv") and \
       os.path.exists("benchmark_outputs/fasttrackpostgreslogging.csv") and \
       os.path.exists("benchmark_outputs/mlflowsqliteretrieval.csv") and \
       os.path.exists("benchmark_outputs/mlflowpostgresretrieval.csv") and \
       os.path.exists("benchmark_outputs/fasttrackpostgresretrieval.csv") and \
       os.path.exists("benchmark_outputs/fasttracksqliteretrieval.csv"):
           return True

    return False


def cleanGeneratedFiles():
    """
    Delete generated output files
    The function checks if a particular csv report output file exists
    and deletes it
    """
    if os.path.exists("benchmark_outputs/mlflowsqlitelogging.csv"):
        os.remove("benchmark_outputs/mlflowsqlitelogging.csv")
    if os.path.exists("benchmark_outputs/mlflowpostgreslogging.csv"):
        os.remove("benchmark_outputs/mlflowpostgreslogging.csv")
    if os.path.exists("benchmark_outputs/fasttracksqlitelogging.csv"):
        os.remove("benchmark_outputs/fasttracksqlitelogging.csv")
    if os.path.exists("benchmark_outputs/fasttrackpostgreslogging.csv"):
        os.remove("benchmark_outputs/fasttrackpostgreslogging.csv")
    if os.path.exists("benchmark_outputs/mlflowsqliteretrieval.csv"):
        os.remove("benchmark_outputs/mlflowsqliteretrieval.csv")
    if os.path.exists("benchmark_outputs/mlflowpostgresretrieval.csv"):
        os.remove("benchmark_outputs/mlflowpostgresretrieval.csv")
    if os.path.exists("benchmark_outputs/fasttrackpostgresretrieval.csv"):
        os.remove("benchmark_outputs/fasttrackpostgresretrieval.csv")
    if os.path.exists("benchmark_outputs/fasttracksqliteretrieval.csv"):
        os.remove("benchmark_outputs/fasttracksqliteretrieval.csv")

    return False


if __name__ == "__main__":
    # ensure all reports have been generated

    logging.info("Beginning report generation")

    # get arguments to python script
    parser = argparse.ArgumentParser()

    parser.add_argument("--clean", help="clean generated csv files after report generation", default=True)
    parser.add_argument(
        "--output", help="the name of the output image, should be a .png file type", default="performanceReport.png"
    )
    parser.add_argument(
        "--numchecks",
        help="the number of times the report generator should check that the csv files have been generated",
        default=10,
    )
    parser.add_argument("--delaybetween", help="the amount of time delay in seconds between checks", default=60)

    args = parser.parse_args()

    OUTPUT_FILE = args.output
    SHOULD_CLEAN = args.clean
    NUM_OF_TIMES_TO_CHECK = args.numchecks
    DELAY_BETWEEN_CHECKS = args.delaybetween

    num_checks = 0

    # while checkAllFilesReady() == False and num_checks < NUM_OF_TIMES_TO_CHECK:
    #     logging.info("Waiting for all csv files to be generated...")
    #     time.sleep(DELAY_BETWEEN_CHECKS)

    if checkAllFilesReady() == True:
        # clean the reports and get the relevant dataframes for the tests
        # generate report using dataframes
        dfs = generateDataframes()
        generateReport(dfs, filename=OUTPUT_FILE)
        logging.info("Report generated successfully")
    else:
        logging.info("Generated CSV files not complete and could not generate reports")


    if SHOULD_CLEAN:
        # cleanGeneratedFiles()
        pass

        cleanGeneratedFiles()
