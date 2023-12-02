import argparse
import logging
import os
import time

import matplotlib.pyplot as plt
import pandas as pd

BENCHMARKS = ['SearchRuns', 'SearchExperiments', 'MetricHistory', 'CreateRun', 
              'LogMetricSingle', 'LogMetricBatch5', 'LogMetricBatch10', 'LogMetricBatch100']

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
)


def generateReport(dfs, filename):
    """
    Generate an image report for a given dataframe and storing it with the 
    provided filename.
    """
    # Sample data
    colors1 = ['red', 'green', 'blue', 'orange']
    
    num_dfs = len(dfs)
    num_cols = int(num_dfs/2)

    # Create a figure with subplots
    fig, axes = plt.subplots(nrows=2, ncols=num_cols, figsize=(30, 20))

    # generate bar charts from dataframes
    for i in range(num_dfs):
        dfs[i].plot(kind='bar', x='application', y='metric_value', ax=axes[int(i/num_cols)][i%num_cols], legend=False, color=colors1,)

    # Customize the layout, labels, and title
    plt.subplots_adjust(wspace=0.4)  # Adjust the space between subplots
    
    for i in range(num_dfs):
        axes[int(i/num_cols)][i%num_cols].set_xlabel('Application')
        axes[int(i/num_cols)][i%num_cols].set_ylabel('Milliseconds')
        axes[int(i/num_cols)][i%num_cols].set_title(dfs[i]['name'][0])
    
    # Save the figure to a single image file (e.g., PNG)
    plt.savefig(filename)
    
    
    
    
def getDataframeFromFile(filename, application_name):
    """
    Generate a dataframe object from reading a file 
    and add a column to the dataframe object to represent what application the file is from
    """
    
    # Read the CSV file into a DataFrame
    df = pd.read_csv(filename)

    # Display the DataFrame
    df = df[df['metric_name'] == 'http_req_duration']
    df = df[df['name'].isin(BENCHMARKS)]
    
    # List of columns to keep
    columns_to_keep = ['metric_value', 'name']

    # Drop all columns except the specified ones
    df.drop(df.columns.difference(columns_to_keep), axis=1, inplace=True)
    df['application'] = application_name
    return df

def generateDataframes():
    """
    Generate single dataframe by concatenating the results from the various report files
    Filter the dataframe for only rows with benchmarks we want to measure
    """
    # get all the dataframes from all the geneated files and indicate the relevant applications
    df1 = getDataframeFromFile('benchmark_outputs/mlflowsqlitethrougput.csv', 'mlflow sqlite')
    df2 = getDataframeFromFile('benchmark_outputs/mlflowpostgresthrougput.csv', 'mlflow postgres')
    df3 = getDataframeFromFile('benchmark_outputs/fasttracksqlitethrougput.csv', 'fasttrack sqlite')
    df4 = getDataframeFromFile('benchmark_outputs/fasttrackpostgresthrougput.csv', 'fasttrack postgres')
    df5 = getDataframeFromFile('benchmark_outputs/mlflowsqliteretreival.csv', 'mlflow sqlite')
    df6 = getDataframeFromFile('benchmark_outputs/mlflowpostgresretreival.csv', 'mlflow postgres')
    df7 = getDataframeFromFile('benchmark_outputs/fasttracksqliteretreival.csv', 'fasttrack sqlite')
    df8 = getDataframeFromFile('benchmark_outputs/fasttrackpostgresretreival.csv', 'fasttrack postgres')


    # Read the CSV file into a DataFrame
    df = pd.concat([df1, df2, df3, df4, df5, df6, df7, df8], ignore_index=True)
    dfs = []
    for benchmark in BENCHMARKS:
        benchmark_df = df[df['name'] == benchmark]
        benchmark_df = benchmark_df.groupby('application')['metric_value'].mean().reset_index()
        benchmark_df['name'] = benchmark
        dfs.append(benchmark_df)
    return dfs



def checkAllFilesReady():
    """
    This is used to check whether all the required output files have been generated
    Since the K6 tests would be run inside containers before shutting down, we need to check if their
    execution is complete before starting the report generataiton. 
    The files we are checking to ensure they exist are:
    - mlflowsqlitethrougput.csv
    - mlflowpostgresthrougput.csv
    - fasttracksqlitethrougput.csv
    - fasttrackpostgresthrougput.csv
    - mlflowsqliteretreival.csv
    - mlflowpostgresretreival.csv
    - fasttrackpostgresretreival.csv
    - fasttracksqliteretreival.csv
    """
    if os.path.exists("benchmark_outputs/mlflowsqlitethrougput.csv") and \
       os.path.exists("benchmark_outputs/mlflowpostgresthrougput.csv") and \
       os.path.exists("benchmark_outputs/fasttracksqlitethrougput.csv") and \
       os.path.exists("benchmark_outputs/fasttrackpostgresthrougput.csv") and \
       os.path.exists("benchmark_outputs/mlflowsqliteretreival.csv") and \
       os.path.exists("benchmark_outputs/mlflowpostgresretreival.csv") and \
       os.path.exists("benchmark_outputs/fasttrackpostgresretreival.csv") and \
       os.path.exists("benchmark_outputs/fasttracksqliteretreival.csv"):
           return True
       
    return False

def cleanGeneratedFiles():
    """
    Delete generated output files
    The function checks if a particular csv report output file exists 
    and deletes it
    """
    if os.path.exists("benchmark_outputs/mlflowsqlitethrougput.csv"):
        os.remove("benchmark_outputs/mlflowsqlitethrougput.csv")
    if os.path.exists("benchmark_outputs/mlflowpostgresthrougput.csv"):
        os.remove("benchmark_outputs/mlflowpostgresthrougput.csv")
    if os.path.exists("benchmark_outputs/fasttracksqlitethrougput.csv"):
        os.remove("benchmark_outputs/fasttracksqlitethrougput.csv")
    if os.path.exists("benchmark_outputs/fasttrackpostgresthrougput.csv"):
        os.remove("benchmark_outputs/fasttrackpostgresthrougput.csv")
    if os.path.exists("benchmark_outputs/mlflowsqliteretreival.csv"):
        os.remove("benchmark_outputs/mlflowsqliteretreival.csv")
    if os.path.exists("benchmark_outputs/mlflowpostgresretreival.csv"):
        os.remove("benchmark_outputs/mlflowpostgresretreival.csv")
    if os.path.exists("benchmark_outputs/fasttrackpostgresretreival.csv"):
        os.remove("benchmark_outputs/fasttrackpostgresretreival.csv")
    if os.path.exists("benchmark_outputs/fasttracksqliteretreival.csv"):
        os.remove("benchmark_outputs/fasttracksqliteretreival.csv")
       
    return False
    

if __name__ == '__main__':
    # ensure all reports have been generated
    
    logging.info("Beginning report generation")
    
    # get arguments to python script
    parser = argparse.ArgumentParser()
    
    parser.add_argument('--clean', help='clean generated csv files after report generation', default=True)
    parser.add_argument('--output', help='the name of the output image, should be a .png file type', default="performanceReport.png")
    parser.add_argument('--numchecks', help='the number of times the report generator should check that the csv files have been generated', default=10)
    parser.add_argument('--delaybetween', help='the amout of time delay in seconds between checks', default=60)
    
    
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
        cleanGeneratedFiles()
    