const contactAlert = document.getElementById(elementId);
const contactForm = document.getElementById('contactForm');
contactForm.addEventListener('submit', async function (e) {
    e.preventDefault();

    const form = e.target;
    const formData = new FormData(form);
    const request = new Request(form.action, {
        method: form.method,
        body: JSON.stringify({ username: "example" }),
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
    });
    const data = {
        name: formData.get('name'),
        email: formData.get('email'),
        message: formData.get('message')
    };
    console.log(data);

    try {
        const response = await fetch(request);

        if (response.ok) {
            // Display success message
            contactAlert.classList.remove('hidden');
            contactAlert.classList.remove('bg-red-500');
            contactAlert.classList.add('bg-green-500');
            contactAlert.innerHTML = 'Message sent successfully';
        } else {
            // Handle errors
            const errorData = await response.json();
            contactAlert.classList.remove('hidden');
            contactAlert.classList.remove('bg-green-500');
            contactAlert.classList.add('bg-red-500');
            contactAlert.innerHTML = 'Error: ' + errorData.message;
        }
    } catch (error) {
        console.error('Error submitting form:', error);
        contactAlert.classList.remove('hidden');
        contactAlert.classList.remove('bg-green-500');
        contactAlert.classList.add('bg-red-500');
        contactAlert.innerHTML = 'Error: ' + error.message;
    }
});