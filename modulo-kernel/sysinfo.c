/* sysinfo.c - /proc/sysinfo_so1_<CARNET> */
#include "common.h"

static int sys_show(struct seq_file *m, void *v)
{
    unsigned long total_kb, free_kb, used_kb;
    struct task_struct *task;

    get_meminfo_kb(&total_kb, &free_kb);
    used_kb = total_kb - free_kb;

    seq_printf(m, "{\n");
    seq_printf(m, "  \"mem_total_kb\": %lu,\n", total_kb);
    seq_printf(m, "  \"mem_free_kb\": %lu,\n", free_kb);
    seq_printf(m, "  \"mem_used_kb\": %lu,\n", used_kb);
    seq_printf(m, "  \"processes\": [\n");

    rcu_read_lock();
    for_each_process(task) {
        unsigned long vsz_kb = 0, rss_kb = 0;
        char cmdline[CMDLINE_MAX] = {0};
        struct mm_struct *mm = task->mm;

        get_mem_from_mm(mm, &vsz_kb, &rss_kb);
        read_task_cmdline(task, cmdline, CMDLINE_MAX);

        seq_printf(m,
            "    { \"pid\": %d, \"name\": \"%s\", \"cmdline\": \"%s\", \"vsz_kb\": %lu, \"rss_kb\": %lu, \"mem_pct\": %.2f, \"state\": \"%ld\" },\n",
            task->pid, task->comm, cmdline, vsz_kb, rss_kb, percent_of(rss_kb, total_kb), task->state);
    }
    rcu_read_unlock();

    seq_printf(m, "  ]\n}\n");
    return 0;
}

static int sys_open(struct inode *inode, struct file *file)
{
    return single_open(file, sys_show, NULL);
}

static const struct file_operations sys_fops = {
    .owner = THIS_MODULE,
    .open = sys_open,
    .read = seq_read,
    .llseek = seq_lseek,
    .release = single_release,
};

static int __init sys_init(void)
{
    proc_create(PROC_NAME_SYS, 0, NULL, &sys_fops);
    pr_info("sysinfo module loaded\n");
    return 0;
}

static void __exit sys_exit(void)
{
    remove_proc_entry(PROC_NAME_SYS, NULL);
    pr_info("sysinfo module unloaded\n");
}

MODULE_LICENSE("GPL");
MODULE_AUTHOR("JUAN CARLOS CHOC - Proyecto SO1");
module_init(sys_init);
module_exit(sys_exit);
