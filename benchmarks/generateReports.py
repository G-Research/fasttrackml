import os
import time

import matplotlib.pyplot as plt
import pandas as pd

BENCHMARKS = ['SearchRuns', 'SearchExperiments', 'MetricHistory', 'CreateRun', 
              'LogMetricSingle', 'LogMetricBatch5', 'LogMetricBatch10', 'LogMetricBatch100']

def generateReport(dfs):
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
        axes[int(i/num_cols)][i%num_cols].set_title(BENCHMARKS[i])
    
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
    df = df[df['metric_name'] == 'http_req_duration']
    df = df[df['name'].isin(BENCHMARKS)]
    
    # List of columns to keep
    columns_to_keep = ['metric_name', 'metric_value', 'name', 'timestamp']

    # Drop all columns except the specified ones
    df.drop(df.columns.difference(columns_to_keep), axis=1, inplace=True)
    df['application'] = application_name
    return df

def generateDataframes():
    # get all the dataframes from all the geneated files and indicate the relevant applications
    df1 = getDataframeFromFile('mlflowsqlitethrougput.csv', 'mlflow sqlite')
    df2 = getDataframeFromFile('mlflowpostgresthrougput.csv', 'mlflow postgres')
    df3 = getDataframeFromFile('fasttracksqlitethrougput.csv', 'fasttrack sqlite')
    df4 = getDataframeFromFile('fasttrackpostgresthrougput.csv', 'fasttrack postgres')
    df5 = getDataframeFromFile('mlflowsqliteretreival.csv', 'mlflow sqlite')
    df6 = getDataframeFromFile('mlflowpostgresretreival.csv', 'mlflow postgres')
    df7 = getDataframeFromFile('fasttracksqliteretreival.csv', 'fasttrack sqlite')
    df8 = getDataframeFromFile('fasttrackpostgresretreival.csv', 'fasttrack postgres')


    # Read the CSV file into a DataFrame
    df = pd.concat([df1, df2, df3, df4, df5, df6, df7, df8], ignore_index=True)
    dfs = []
    for benchmark in BENCHMARKS:
        benchmark_df = df[df['name'] == benchmark]
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
    if os.path.exists("mlflowsqlitethrougput.csv") and \
       os.path.exists("mlflowpostgresthrougput.csv") and \
       os.path.exists("fasttracksqlitethrougput.csv") and \
       os.path.exists("fasttrackpostgresthrougput.csv") and \
       os.path.exists("mlflowsqliteretreival.csv") and \
       os.path.exists("mlflowpostgresretreival.csv") and \
       os.path.exists("fasttrackpostgresretreival.csv") and \
       os.path.exists("fasttracksqliteretreival.csv"):
           return True
       
    return False

NUM_OF_TIMES_TO_CHECK = 10
DELAY_BETWEEN_CHECKS = 60

if __name__ == '__main__':
    # ensure all reports have been generated
    num_checks = 0
    while checkAllFilesReady() == False and num_checks < NUM_OF_TIMES_TO_CHECK:
        time.sleep(DELAY_BETWEEN_CHECKS)
    
    if checkAllFilesReady() == True:
        # clean the reports and get the relevant dataframes for the tests
        # generate report using dataframes
        dfs = generateDataframes()
        generateReport(dfs)
    else:
        print("Generated CSV files not complete and could not generate reports")
    