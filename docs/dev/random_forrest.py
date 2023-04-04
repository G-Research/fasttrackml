import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split

import mlflow

ALL_FILES = ["data/0.csv", "data/1.csv", "data/2.csv", "data/3.csv"]

mlflow.set_tracking_uri("http://localhost:5000/mlflow")
mlflow.set_experiment("my-experiment")

mlflow.sklearn.autolog()


def read_data():
    """
    Read the data from the csv files and return the dataframes
    and the train and test sets.
    The Ratio of the train and test sets is 50:50.
    """
    array = []

    for file in ALL_FILES:
        read = pd.read_csv(file, header=None)
        array.append(read)

    data_df = pd.concat(array)
    X = data_df.iloc[:, :-1].values
    Y = data_df.iloc[:, -1].values
    # Random State is set to 42 for reproducibility
    Xtrain, Xtest, Ytrain, Ytest = train_test_split(
        X, Y, test_size=0.50, random_state=42
    )

    print("Completed Data Read")
    return data_df, Xtrain, Xtest, Ytrain, Ytest


def hyper_tune(X_train_s, y_train):
    """
    Tune the hyperparameters of the Random Forest Classifier
    using Out of Bag Error.

    Here we specifically tune the number of trees and the
    max_features parameter.
    """

    oob_dfs = []
    for temp_max_features in ["sqrt", "log2", None]:
        # Max Depth was set to 3 to reduce the number of trees
        # required to achieve a good OOB score.
        # Random State is set to 42 for reproducibility
        rf_test = RandomForestClassifier(
            oob_score=True,
            random_state=42,
            max_features=temp_max_features,
            max_depth=3,
            n_jobs=-1,
        )
        oob_list = []
        for n_trees in range(50, 500, 50):
            rf_test.set_params(n_estimators=n_trees)
            rf_test.fit(X_train_s, y_train)
            oob_error = 1 - rf_test.oob_score_
            oob_list.append(pd.Series({"n_trees": n_trees, "oob": oob_error}))

        oob_df = pd.concat(oob_list, axis=1).T.set_index("n_trees")
        oob_dfs.append((temp_max_features, oob_df))

    print("Completed Hyperparameter Tuning")
    return oob_dfs


def main():
    _, X_train, X_test, y_train, _ = read_data()

    rf_oob_dfs = hyper_tune(X_train, y_train)

    # Find the best max_features and best tree count
    # by looking at the minimum OOB error
    min_oob = min([rf_oob_df["oob"].min() for max_features, rf_oob_df in rf_oob_dfs])
    for max_features, rf_oob_df in rf_oob_dfs:
        if rf_oob_df["oob"].min() == min_oob:
            BEST_MAX_FEATURES = max_features
            BEST_TREE_COUNT = rf_oob_df["oob"].idxmin()
            break

    # Random State is set to 42 for reproducibility
    optimised_classifier = RandomForestClassifier(
        n_estimators=int(BEST_TREE_COUNT),
        random_state=42,
        n_jobs=-1,
        max_features=BEST_MAX_FEATURES,
    )

    optimised_classifier.fit(X_train, y_train)

    optimised_classifier.predict(X_test)

    print("Made Predictions")


if __name__ == "__main__":
    main()
