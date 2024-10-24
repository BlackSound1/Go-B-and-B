/**
 * Prompt is a utility function that provides various types of alert dialogs
 * using SweetAlert2. It offers different methods to show toast notifications,
 * success dialogs, error dialogs, and custom dialogs with configurable options.
 *
 * Methods:
 * - toast(c): Displays a toast notification with configurable message, icon,
 *   and position.
 * - success(c): Shows a success dialog with a message, title, and footer.
 * - error(c): Displays an error dialog with a message, title, and footer.
 * - custom(c): Presents a customizable dialog with options for icon, message,
 *   title, confirmation buttons, and optional callbacks for dialog events.
 *
 * @returns {Object} An object containing methods: toast, success, error, and custom.
 */
function Prompt() {

    /**
     * Displays a toast notification with configurable message, icon, and position.
     *
     * @param {Object} c Options for the notification.
     * @param {string} [c.msg=''] Message to be displayed.
     * @param {string} [c.icon='success'] Icon to be used.
     * @param {string} [c.position='top-end'] Position of the notification.
     * @returns {void}
     */
    function toast(c) {
        // Destructure the options
        const{
            msg = '',
            icon = 'success',
            position = 'top-end',

        } = c;

        // Create the toast using the options
        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer);
                toast.addEventListener('mouseleave', Swal.resumeTimer);
            }
        });

        // Fire off the toast
        Toast.fire({});
    };

    /**
     * Displays a success notification with configurable title, message, and footer.
     *
     * @param {Object} c Options for the notification.
     * @param {string} [c.msg=''] Message to be displayed.
     * @param {string} [c.title=''] Title of the notification.
     * @param {string} [c.footer=''] Footer of the notification.
     * @returns {void}
     */
    function success(c) {
        // Destructure the options
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;

        // Fire off the success notification wiith the options
        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer,
        });
    };

    /**
     * Displays an error notification with configurable title, message, and footer.
     *
     * @param {Object} c Options for the notification.
     * @param {string} [c.msg=''] Message to be displayed.
     * @param {string} [c.title=''] Title of the notification.
     * @param {string} [c.footer=''] Footer of the notification.
     * @returns {void}
     */
    function error(c) {
        // Destructure the options
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;

        // Fire off the error notification wiith the options
        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer,
        });
    };

    /**
     * Displays a custom notification with configurable title, message, and footer.
     *
     * @param {Object} c Options for the notification.
     * @param {string} [c.icon=''] Icon of the notification.
     * @param {string} [c.msg=''] Message to be displayed.
     * @param {string} [c.title=''] Title of the notification.
     * @param {boolean} [c.showConfirmButton=true] Show the confirm button.
     * @param {function} [c.callback=undefined] Callback function to be called when the notification is closed.
     * @param {function} [c.willOpen=undefined] Callback function to be called when the notification is about to open.
     * @param {function} [c.didOpen=undefined] Callback function to be called when the notification is opened.
     * @returns {void}
     */
    async function custom(c) {
        // Destructure the options
        const {
            icon = "",
            msg = "",
            title = "",
            showConfirmButton = true
        } = c;

        // Fire off the custom notification with the options, saving the result
        const { value: result } = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            willOpen: () => {
                // If a willOpen callback is defined, call it
                if (c.willOpen !== undefined) {
                    c.willOpen();
                }
            },
            didOpen: () => {
                // If a didOpen callback is defined, call it
                if (c.didOpen !== undefined) {
                    c.didOpen();
                }
            },
            preConfirm: () => {
                return [
                    document.getElementById('start').value,
                    document.getElementById('end').value
                ];
            }
        });

        // If there is a result
        if (result) {

            // If the result is not cancel
            if (result.dismiss !== Swal.DismissReason.cancel) {

                // If the result is not empty
                if (result.value !== "") {

                    // If there is a callback
                    if (c.callback !== undefined) {
                        c.callback(result);
                    }
                } else {
                    c.callback(false);
                }
            }
        } else {
            c.callback(false);
        }
    };

    // Return the methods
    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    };
}

/**
 * Handles the event listener for the "Check Availability" button on the rooms page.
 * Creates a custom modal for the user to choose the dates, and then sends a POST
 * request to the server to check for availability. If there is availability, it
 * displays a success notification with a link to book the room. If there is no
 * availability, it displays an error notification.
 *
 * @param {number} room_id The room ID to check availability for.
 * @param {string} token The CSRF token to add to the POST request.
 */
function HandleBookingOnRoomsPage(room_id, token) {

    document.getElementById("check-availability-button").addEventListener("click", () => {
        let html = `
            <form id="check-availability-form" action="" method="post" novalidate class="needs-validation">
                <div class="form-row">
                    <div class="col">
                        <div class="form-row" id="reservation-dates-modal">
                            <div class="col">
                                <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">
                            </div>

                            <div class="col">
                                <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                            </div>
                        </div>
                    </div>
                </div>
            </form>
        `;

        // Create a custom modal
        attention.custom({ 
            title: 'Choose your dates', 
            msg: html,
            willOpen: () => {
                // Before modal opens

                const elem = document.getElementById("reservation-dates-modal");
                const rp = new DateRangePicker(elem, {
                    format: 'yyyy-mm-dd',
                    showOnFocus: true,
                    minDate: new Date(),
                });
            },
            didOpen: () => {
                // After modal opens

                document.getElementById("start").removeAttribute("disabled");
                document.getElementById("end").removeAttribute("disabled");
            },
            callback: (result) => {
                // Once the modal is closed

                let form = document.getElementById("check-availability-form");
                let formData = new FormData(form);
                formData.append("csrf_token", token); // Add the CSRF token dynamically
                formData.append("room_id", room_id);

                // Get the data, convert it to JSON, and POST it
                fetch("/search-availability-json", {
                    method: "POST",
                    body: formData
                })
                .then(response => response.json())
                .then(data => {
                    if (data.ok) {
                        attention.custom({
                            icon: "success",
                            showConfirmButton: false,
                            msg: '<p>Room is available</p>'
                                 + '<p><a href="/book-room?id=' 
                                 + data.room_id 
                                 + '&s=' 
                                 + data.start_date 
                                 + '&e=' 
                                 + data.end_date 
                                 + '" class="btn btn-primary">Book Now</a></p>'
                        });
                    } else {
                        attention.error({ msg: "No availability" });
                    }
                });
            }
        });
    });
}
