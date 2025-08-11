document.addEventListener('DOMContentLoaded', () => {
    const searchBtn = document.getElementById('searchBtn');
    const orderIdInput = document.getElementById('orderId');
    const resultDiv = document.getElementById('result');
    
    searchBtn.addEventListener('click', getOrder);
    orderIdInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            getOrder();
        }
    });

    async function getOrder() {
        const orderId = orderIdInput.value.trim();
        if (!orderId) {
            alert('Please enter Order ID');
            return;
        }

        try {
            resultDiv.textContent = 'Loading...';
            const response = await fetch(`/order/${encodeURIComponent(orderId)}`);
            
            if (!response.ok) {
                throw new Error(await response.text());
            }
            
            const data = await response.json();
            resultDiv.textContent = JSON.stringify(data, null, 2);
        } catch (error) {
            resultDiv.textContent = `Error: ${error.message}`;
        }
    }
});