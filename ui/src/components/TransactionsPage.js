import { TransactionsTable } from "./TransactionsTable";

export default function TransactionsPage() {

    var txnList = []

    // Fetch the stock list from the server.
    function serverRequest() {
        console.log('Fetching transaction data...')
        fetch("http://localhost:5000/transactions")
            .then(response => response.json())
            .then(resp => txnList = resp["transactions"])
    }

    // Fetch the transaction data from the server.
    serverRequest();

    return (
        <>
            {/* Display all the transactions in a sortable table. */}
            <TransactionsTable
                txnData={txnList}
            />
        </>
    );
}