#include "common.h"
#include <linux/sched.h>
#include <linux/sched/task.h>
#include <linux/sched/signal.h>


static bool is_container_task(char *cmdline)
{
    if (!cmdline)
        return false;

    if (strstr(cmdline, "docker")) return true;
    if (strstr(cmdline, "container")) return true;
    if (strstr(cmdline, "runc")) return true;
    if (strstr(cmdline, "busybox")) return true;

    return false;
}

static int cont_show(struct seq_file *m, void *v)
{
    struct task_struct *task;
    unsigned long total_kb, free_kb, used_kb;

    get_meminfo_kb(&total_kb, &free_kb);
    used_kb = total_kb - free_kb;

    seq_printf(m, "{\n");
    seq_printf(m, "  \"mem_total_kb\": %lu,\n", total_kb);
    seq_printf(m, "  \"mem_free_kb\": %lu,\n", free_kb);
    seq_printf(m, "  \"mem_used_kb\": %lu,\n", used_kb);
    seq_printf(m, "  \"containers\": [\n");

    rcu_read_lock();
    for_each_process(task) {

        unsigned long vsz_kb = 0, rss_kb = 0;
        char cmdline[CMDLINE_MAX] = {0};

        /* Obtener cmdline */
        read_task_cmdline(task, cmdline, CMDLINE_MAX);

        /* Filtrar procesos que parecen ser contenedores */
        if (!is_container_task(cmdline))
            continue;

        /* Obtener vsz y rss */
        get_mem_from_mm(task->mm, &vsz_kb, &rss_kb);

        /* calcular porcentaje sin float */
        unsigned long pct_x100 = percent_of_x100(vsz_kb, total_kb);
        unsigned long pct_int  = pct_x100 / 100;
        unsigned long pct_dec  = pct_x100 % 100;

        seq_printf(m,
            "    { \"pid\": %d, \"name\": \"%s\", \"cmdline\": \"%s\", "
            "\"vsz_kb\": %lu, \"rss_kb\": %lu, "
            "\"mem_pct\": \"%lu.%02lu\" },\n",
            task->pid, task->comm, cmdline,
            vsz_kb, rss_kb,
            pct_int, pct_dec
        );
    }
    rcu_read_unlock();

    seq_printf(m, "  ]\n}\n");
    return 0;
}

static int cont_open(struct inode *inode, struct file *file)
{
    return single_open(file, cont_show, NULL);
}

/* proc_ops para kernel 6.x */
static const struct proc_ops cont_fops = {
    .proc_open    = cont_open,
    .proc_read    = seq_read,
    .proc_lseek   = seq_lseek,
    .proc_release = single_release,
};

static int __init cont_init(void)
{
    proc_create(PROC_NAME_CONT, 0, NULL, &cont_fops);
    pr_info("continfo module loaded\n");
    return 0;
}

static void __exit cont_exit(void)
{
    remove_proc_entry(PROC_NAME_CONT, NULL);
    pr_info("continfo module unloaded\n");
}

MODULE_LICENSE("GPL");
MODULE_AUTHOR("JUAN CARLOS CHOC - Proyecto SO1");
module_init(cont_init);
module_exit(cont_exit);
