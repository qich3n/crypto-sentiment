let twitterEnabled = false;

// Check if Twitter is enabled on page load
async function checkTwitterStatus() {
    try {
        const response = await fetch('/api/v1/health');
        const data = await response.json();
        twitterEnabled = data.services.twitter;
    } catch (error) {
        console.error('Error checking Twitter status:', error);
        twitterEnabled = false; // Assume Twitter is disabled if the check fails > remove or not
    }
}

// Analyze sentiment
async function analyzeSentiment() {
    const symbol = document.getElementById('symbol').value.trim();
    const resultDiv = document.getElementById('result');

    // Clear previous results
    resultDiv.innerHTML = '';

    if (!symbol) {
        resultDiv.innerHTML = `<div class="text-red-500">Please enter a valid cryptocurrency symbol.</div>`;
        return;
    }

    // Dynamically display "Analyzing..." message
    resultDiv.innerHTML = `<div class="text-blue-500">Analyzing...</div>`;

    try {
        const response = await fetch(`/api/v1/sentiment/${symbol}`);
        if (!response.ok) throw new Error('Failed to fetch sentiment data');
        const data = await response.json();

        // Display result
        resultDiv.innerHTML = `
            <div class="bg-gray-50 p-4 rounded">
                <div class="text-xl font-bold mb-4">
                    ${data.symbol.toUpperCase()}: ${(data.overall_score * 100).toFixed(1)}% 
                    ${data.overall_score > 0.1 ? '📈 Bullish' : data.overall_score < -0.1 ? '📉 Bearish' : '↔️ Neutral'}
                </div>
                <div class="mb-2">Reddit Score: ${(data.reddit_score * 100).toFixed(1)}% (${data.reddit_posts} posts)</div>
                ${
                    twitterEnabled
                        ? `<div>Twitter Score: ${(data.twitter_score * 100).toFixed(1)}% (${data.tweets} tweets)</div>`
                        : ''
                }
            </div>
        `;
    } catch (error) {
        console.error('Error fetching sentiment data:', error);
        resultDiv.innerHTML = `<div class="text-red-500">Error fetching data. Please try again.</div>`;
    }
}

// Check Twitter status when page loads
document.addEventListener('DOMContentLoaded', checkTwitterStatus);
