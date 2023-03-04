import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";

// Fetch the transaction data from the server, and return/render the transactions page.
export default function TransactionsPage() {
    const [txnList, setTxnList] = useState([]);

    document.body.style.backgroundColor = "black"

    // Fetch the transaction data from the server.
    useEffect(() => {
        console.log('Fetching transaction data...')
        fetch("http://localhost:5000/transactions")
            .then(response => response.json())
            .then(resp => setTxnList(resp["transactions"]))
    }, []);

    return (
        <>
            {/* Display all the transactions in a sortable table. */}
            <TransactionsTable
                txnData={txnList.filter(t => (t.action === "Buy" || t.action === "Sell"))}
            />
        </>
    );
}