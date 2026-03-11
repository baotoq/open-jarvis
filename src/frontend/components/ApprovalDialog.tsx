'use client'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface ApprovalDialogProps {
  open: boolean
  command: string
  onApprove: () => void
  onDeny: () => void
}

export function ApprovalDialog({ open, command, onApprove, onDeny }: ApprovalDialogProps) {
  return (
    <Dialog open={open}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Shell Command Approval</DialogTitle>
          <DialogDescription>The agent wants to run this command:</DialogDescription>
        </DialogHeader>
        <pre className="rounded-md bg-muted px-4 py-3 font-mono text-sm overflow-auto">{command}</pre>
        <DialogFooter>
          <Button variant="destructive" onClick={onDeny}>Deny</Button>
          <Button onClick={onApprove}>Allow</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
