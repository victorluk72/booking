
 // Prompt is a JS module for all alerts, notifications and custom pop up dialogs
function Prompt(){

    let success = function(c){
        const {
            msg ="",
            icon ="success",
            position = "top-end",

        }=c;

        const Toast = Swal.mixin({
              toast: true,
              title: msg,
              position: position,
              icon:icon,
              showConfirmButton: false,
              timer: 2000,
              timerProgressBar: true,
              didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
              }
     })

        Toast.fire({})
    }

    let error = function(c){
        const {
            msg ="",
            icon ="error",
            position = "top-end",

        }=c;

        const Toast = Swal.mixin({
              toast: true,
              title: msg,
              position: position,
              icon:icon,
              showConfirmButton: false,
              timer: 2000,
              timerProgressBar: true,
              didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
              }
     })

        Toast.fire({})
        
    }

    async function custom(c){
        const {
            icon="",
            msg ="",
            title ="",
            showConfirmButton = true,
        }=c;

    const { value: result } = await Swal.fire({
              icon:icon,
              title: title,
              html:msg,
              backdrop: false,
              focusConfirm: false,
              showCancelButton: true,
              showConfirmButton: showConfirmButton, 
              willOpen: () => {
                  if (c.willOpen !== undefined){
                      c.willOpen();
                  }                                   
              },
              didOpen: () => {
                if (c.didOpen !== undefined){
                      c.didOpen();
                  }
                }
            })

            //----This block is to handle result of date pick (asyncronically)
            //Check if you have any result and handle the result
            if (result){

                //Check if user didin't click "Cancel" button
                if (result.dismiss !== Swal.DismissReason.cancel){

                    //Check if value of result is not empty
                    if (result.value !==""){
                        
                        if (c.callback !== undefined){
                            //pass result asyncronically 
                            c.callback(result);
                        }
                    }else{
                        c.callback(false);
                    }

                //This else if result is "Cancel"
                }else{
                    c.callback(false);
                }
            }
            //----End of the block that handles result of date pick (asyncronically)
        }
        

   return {
    success: success,
    error:error,
    custom:custom,
   }
}