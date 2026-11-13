'use server'

export async function submitAction(formData) {
  const input = formData.get('input')
  return { message: "Received: " + input }
}
