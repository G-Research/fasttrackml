import mlflow


class MlflowWrap:
    def __getattr__(self, attr):
        return getattr(mlflow, attr)

    def custom_function1(self):
        print("custom1")

    def custom_function2(self):
        # Tua logica personalizzata per la seconda funzione
        print("custom2")

mlflow_wrap = MlflowWrap()